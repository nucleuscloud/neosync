package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/userconfig"
)

// Attempts to find the account id from various places in the following order:
// 1. If accountIdFlag is not empty, uses that
// 2. If API is provided, attempts to resolve the accountid from it
// 3. Checks the account context for a stored account id.
// 4. Fail
func ResolveAccountIdFromFlag(
	ctx context.Context,
	userclient mgmtv1alpha1connect.UserAccountServiceClient,
	accountIdFlag *string,
	apiKey *string,
	logger *slog.Logger,
) (string, error) {
	if accountIdFlag != nil && *accountIdFlag != "" {
		logger.Debug(fmt.Sprintf("provided account id %q set from flag", *accountIdFlag))
		return *accountIdFlag, nil
	}
	if apiKey != nil && *apiKey != "" {
		logger.Debug("api key detected, attempting to resolve account id from key.")
		uaResp, err := userclient.GetUserAccounts(
			ctx,
			connect.NewRequest(&mgmtv1alpha1.GetUserAccountsRequest{}),
		)
		if err != nil {
			return "", fmt.Errorf("unable to resolve account id from api key: %w", err)
		}
		apiKeyAccounts := uaResp.Msg.GetAccounts()
		if len(apiKeyAccounts) == 0 {
			return "", errors.New("api key is not associated with any neosync accounts")
		}
		accountId := apiKeyAccounts[0].GetId()
		logger.Debug(fmt.Sprintf("provided api key resolved to account %q", accountId))
		return accountId, nil
	}
	accountId, err := userconfig.GetAccountId()
	if err != nil {
		return "", fmt.Errorf(
			`unable to resolve account id from account context, please use the "neosync accounts switch" command to set an active account context: %w`,
			err,
		)
	}
	logger.Debug(fmt.Sprintf("account id %q resolved from user config", accountId))
	return accountId, nil
}
