package login_cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	cli_logger "github.com/nucleuscloud/neosync/cli/internal/logger"
	"github.com/nucleuscloud/neosync/cli/internal/userconfig"
	"github.com/nucleuscloud/neosync/cli/internal/version"
	http_client "github.com/nucleuscloud/neosync/internal/http/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/toqueteos/webbrowser"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Neosync",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}
			debugMode, err := cmd.Flags().GetBool("debug")
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			logger := cli_logger.NewSLogger(cli_logger.GetCharmLevelOrDefault(debugMode))

			if apiKey != "" {
				logger.Info(
					`found api key, no need to log in. run "neosync whoami" to verify that the api key is valid`,
				)
				return nil
			}
			return login(cmd.Context(), logger)
		},
	}
	return cmd
}

func login(ctx context.Context, logger *slog.Logger) error {
	httpclient := http_client.NewWithHeaders(version.Get().Headers())
	authclient := mgmtv1alpha1connect.NewAuthServiceClient(httpclient, auth.GetNeosyncUrl())
	isAuthEnabled, err := auth.GetAuthEnabled(ctx, authclient)
	if err != nil {
		return err
	}
	if !isAuthEnabled {
		logger.Info("auth is not enabled server-side, exiting")
		return nil
	}
	err = oAuthLogin(ctx, authclient)
	if err != nil {
		return err
	}

	httpclient, err = auth.GetNeosyncHttpClient(ctx, logger)
	if err != nil {
		return err
	}

	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(
		httpclient,
		auth.GetNeosyncUrl(),
	)

	accountsResp, err := userclient.GetUserAccounts(
		ctx,
		connect.NewRequest(&mgmtv1alpha1.GetUserAccountsRequest{}),
	)
	if err != nil {
		return err
	}

	for _, a := range accountsResp.Msg.Accounts {
		if strings.EqualFold(a.Name, "personal") {
			// set account to personal
			err = userconfig.SetAccountId(a.Id)
			if err != nil {
				return fmt.Errorf("unable to set account id in context: %w", err)
			}
			logger.Info(fmt.Sprintf("  Account set to %q", a.Name))
		}
	}
	return nil
}

const (
	callbackPath = "/api/auth/callback"
)

var (
	redirectUri = fmt.Sprintf("http://%s%s", getRedirectUriBaseUrl(), callbackPath)
)

type oauthResult struct {
	Code  string
	State string
}

func oAuthLogin(
	ctx context.Context,
	authclient mgmtv1alpha1connect.AuthServiceClient,
) error {
	state := uuid.NewString()

	authorizeurlResp, err := authclient.GetAuthorizeUrl(
		ctx,
		connect.NewRequest(&mgmtv1alpha1.GetAuthorizeUrlRequest{
			State:       state,
			RedirectUri: redirectUri,
			Scope:       "openid profile offline_access",
		}),
	)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	errchan := make(chan error)
	reschan := make(chan oauthResult)

	mux.HandleFunc(callbackPath, getHttpCallbackFunc(state, errchan, reschan))
	httpSrv := http.Server{
		Addr:              getHttpSrvBaseUrl(),
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ErrorLog:          nil,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			errchan <- err
		}
	}()

	if err := webbrowser.Open(authorizeurlResp.Msg.Url); err != nil {
		fmt.Println( //nolint:forbidigo
			"There was an issue opening the web browser, proceed to the following url to finish logging in to Neosync:\n",
			authorizeurlResp.Msg.Url,
		)
	}

	select {
	case <-time.After(5 * time.Minute):
		close(errchan)
		close(reschan)
		return errors.New("timed out waiting for client")
	case err := <-errchan:
		close(errchan)
		close(reschan)
		return err
	case result := <-reschan:
		close(errchan)
		close(reschan)
		if result.State != state {
			return errors.New("state received from response was not what was sent")
		}
		loginResp, err := authclient.LoginCli(
			ctx,
			connect.NewRequest(
				&mgmtv1alpha1.LoginCliRequest{Code: result.Code, RedirectUri: redirectUri},
			),
		)
		if err != nil {
			return err
		}
		err = userconfig.SetAccessToken(loginResp.Msg.AccessToken.AccessToken)
		if err != nil {
			return err
		}
		if loginResp.Msg.AccessToken.RefreshToken != nil {
			err = userconfig.SetRefreshToken(*loginResp.Msg.AccessToken.RefreshToken)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func getHttpCallbackFunc(
	state string,
	errchan chan<- error,
	reschan chan<- oauthResult,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		queryResp := getCallbackResponse(r.URL.Query())
		if queryResp.Error != "" || queryResp.ErrorDescription != "" {
			if err := renderLoginErrorPage(w, loginPageErrorData{
				Title:            "Login Failed",
				ErrorCode:        queryResp.Error,
				ErrorDescription: queryResp.ErrorDescription,
			}); err != nil {
				errchan <- err
				return
			}
			errchan <- errors.New("unable to finish login flow")
			return
		}
		if queryResp.Code == "" || queryResp.State == "" {
			if err := renderLoginErrorPage(w, loginPageErrorData{
				Title:            "Login Failed",
				ErrorCode:        "BadRequest",
				ErrorDescription: "Missing required query parameters to finish logging in.",
			}); err != nil {
				errchan <- err
				return
			}
			errchan <- errors.New("received invalid callback response")
			return
		}
		if state != queryResp.State {
			if err := renderLoginErrorPage(w, loginPageErrorData{
				Title:            "Login Failed",
				ErrorCode:        "BadRequest",
				ErrorDescription: "Received invalid state in response",
			}); err != nil {
				errchan <- err
				return
			}
			errchan <- errors.New("received invalid state in response")
			return
		}
		if err := renderLoginSuccessPage(w, loginPageData{Title: "Login Success!"}); err != nil {
			errchan <- err
			return
		}
		reschan <- oauthResult{Code: queryResp.Code, State: queryResp.State}
	}
}

type callbackResponseQuery struct {
	Code             string
	State            string
	Error            string
	ErrorDescription string
}

func getCallbackResponse(
	query url.Values,
) callbackResponseQuery {
	return callbackResponseQuery{
		Code:             query.Get("code"),
		State:            query.Get("state"),
		Error:            query.Get("error"),
		ErrorDescription: query.Get("error_description"),
	}
}
func getRedirectUriBaseUrl() string {
	return fmt.Sprintf("%s:%d", getHttpRedirectHost(), getHttpSrvPort())
}
func getHttpSrvBaseUrl() string {
	return fmt.Sprintf("%s:%d", getHttpSrvHost(), getHttpSrvPort())
}
func getHttpSrvHost() string {
	host := viper.GetString("LOGIN_HOST")
	if host == "" {
		return "127.0.0.1"
	}
	return host
}

func getHttpRedirectHost() string {
	host := viper.GetString("LOGIN_REDIRECT_HOST")
	if host == "" {
		return "127.0.0.1"
	}
	return host
}

func getHttpSrvPort() uint32 {
	port := viper.GetUint32("LOGIN_PORT")
	if port == 0 {
		return 4242
	}
	return port
}
