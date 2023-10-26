package login_cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
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
			cmd.SilenceUsage = true
			return login(cmd.Context())
		},
	}
	return cmd
}

func login(ctx context.Context) error {
	authclient := mgmtv1alpha1connect.NewAuthServiceClient(http.DefaultClient, "")
	authEnabledResp, err := authclient.GetAuthStatus(ctx, connect.NewRequest(&mgmtv1alpha1.GetAuthStatusRequest{}))
	if err != nil {
		return err
	}
	if !authEnabledResp.Msg.IsEnabled {
		fmt.Println("auth is not enabled server-side, exiting")
		return nil
	}
	return oAuthLogin(ctx, authclient)
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

	authorizeurlResp, err := authclient.GetAuthorizeUrl(ctx, connect.NewRequest(&mgmtv1alpha1.GetAuthorizeUrlRequest{
		State:       state,
		RedirectUri: redirectUri,
	}))
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
		fmt.Println("There was an issue opening the web browser, proceed to the following url to finish logging in to Neosync:\n", authorizeurlResp.Msg.Url)
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
		loginResp, err := authclient.LoginCli(ctx, connect.NewRequest(&mgmtv1alpha1.LoginCliRequest{Code: result.Code, RedirectUri: redirectUri}))
		if err != nil {
			return err
		}
		fmt.Println("AccessToken", loginResp.Msg.AccessToken.AccessToken)
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
