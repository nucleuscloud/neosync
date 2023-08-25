package authvalidate

import (
	"context"
)

// type accountContextKey struct{}
// type AccountContextData struct {
// 	Account *nucleusdb.Account
// }

type UserContextKey struct{}
type UserContextData struct {
	UserId                 string
	IsNucleusOwned         bool
	ServiceAccountClientId string
}

// func InjectAuthAccountCtx(ctx context.Context, db nucleusdb.NucleusDbInterface) (context.Context, error) {
// 	// logger, err := loggermiddleware.GetLoggerFromContext(ctx)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	tokenCtxData, err := authjwt.GetTokenDataFromCtx(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if tokenCtxData == nil {
// 		return nil, nucleuserrors.NewUnauthenticated("unable to find token context data")
// 	}

// 	if tokenCtxData.IsServiceAccount && tokenCtxData.IsNucleusOwned {
// 		fields := grpc_logger.Fields{"IsNucleusOwned", strconv.FormatBool(tokenCtxData.IsNucleusOwned), "ServiceAccountClientId", tokenCtxData.AuthUserId}
// 		ctx = loggermiddleware.AddFieldsToLogger(fields, ctx, logger)
// 		return ctx, nil
// 	} else if tokenCtxData.IsServiceAccount {
// 		var account *nucleusdb.Account
// 		account, err = db.GetAccountByServiceAccountClientId(ctx, tokenCtxData.AuthUserId)
// 		if err != nil {
// 			return nil, err
// 		}
// 		fields := grpc_logger.Fields{"ServiceAccountClientId", tokenCtxData.AuthUserId, "accountId", account.Id}
// 		ctx = context.WithValue(ctx, UserContextKey{}, &UserContextData{
// 			IsNucleusOwned:         false,
// 			ServiceAccountClientId: tokenCtxData.AuthUserId,
// 		})
// 		ctx = loggermiddleware.AddFieldsToLogger(fields, ctx, logger)
// 		return context.WithValue(ctx, accountContextKey{}, &AccountContextData{
// 			Account: account,
// 		}), nil
// 	}

// 	user, err := db.GetUserAccount(ctx, tokenCtxData.AuthUserId)
// 	if err != nil {
// 		logger.Error(err, "unable to find user record")
// 		if err == sql.ErrNoRows {
// 			return nil, nucleuserrors.NewUnauthenticated("unable to find valid user record")
// 		}
// 		return nil, nucleuserrors.NewInternalError("unable to check for valid user record")
// 	}

// 	ctx = context.WithValue(ctx, UserContextKey{}, &UserContextData{
// 		UserId:         user.UserId,
// 		IsNucleusOwned: false,
// 	})
// 	// Standard user
// 	if tokenCtxData.HasOrganization() {
// 		// get org account
// 		account, err := db.GetOrgAccount(ctx, user.UserId, *tokenCtxData.OrganizationId)
// 		if err != nil {
// 			logger.Error(err, "unable to find org user account", "userId", user.UserId)
// 			if err == sql.ErrNoRows {
// 				return nil, nucleuserrors.NewUnauthenticated("unable to find valid org account associated with this user")
// 			}
// 			return nil, nucleuserrors.NewInternalError("unable to check for valid org account associated with this user")
// 		}
// 		fields := grpc_logger.Fields{"UserId", user.UserId, "accountId", account.Id}
// 		ctx = loggermiddleware.AddFieldsToLogger(fields, ctx, logger)
// 		return context.WithValue(ctx, accountContextKey{}, &AccountContextData{
// 			Account: account,
// 		}), nil

// 	} else {
// 		// get personal account
// 		account, err := db.GetPersonalAccount(ctx, user.UserId)
// 		if err != nil {
// 			logger.Error(err, "unable to find personal user account", "userId", user.UserId)
// 			if err == sql.ErrNoRows {
// 				return nil, nucleuserrors.NewUnauthenticated("unable to find valid personal account associated with this user")
// 			}
// 			return nil, nucleuserrors.NewInternalError("unable to check for valid personal account associated with this user")
// 		}
// 		fields := grpc_logger.Fields{"UserId", user.UserId, "accountId", account.Id}
// 		ctx = loggermiddleware.AddFieldsToLogger(fields, ctx, logger)
// 		return context.WithValue(ctx, accountContextKey{}, &AccountContextData{
// 			Account: account,
// 		}), nil
// 	}
// }

// func getAccountDataFromCtx(ctx context.Context) (*AccountContextData, error) {
// 	data, ok := ctx.Value(accountContextKey{}).(*AccountContextData)
// 	if !ok {
// 		return nil, nucleuserrors.NewInternalError("ctx does not contain accountContextKey or unable to cast struct")
// 	}
// 	return data, nil
// }

// func GetAccountFromCtx(ctx context.Context) (*nucleusdb.Account, error) {
// 	data, err := getAccountDataFromCtx(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return data.Account, nil
// }

// func GetAccountFromCtxStrict(ctx context.Context) (*nucleusdb.Account, error) {
// 	account, err := GetAccountFromCtx(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if account == nil {
// 		return nil, nucleuserrors.NewUnauthenticated("user is not associated with a nucleus account")
// 	}
// 	return account, err
// }

func GetUserDataFromCtx(ctx context.Context) *UserContextData {
	data, ok := ctx.Value(UserContextKey{}).(*UserContextData)
	if !ok {
		return &UserContextData{}
	}
	return data
}
