package v1alpha1_apikeyservice

import (
	"context"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
)

func (s *Service) GetAccountApiKeys(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccountApiKeysRequest],
) (*connect.Response[mgmtv1alpha1.GetAccountApiKeysResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("accountId", req.Msg.AccountId)
	_ = logger

	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	apiKeys, err := s.db.Q.GetAccountApiKeys(ctx, s.db.Db, *accountUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.GetAccountApiKeysResponse{
			ApiKeys: []*mgmtv1alpha1.AccountApiKey{},
		}), nil
	}

	dtos := make([]*mgmtv1alpha1.AccountApiKey, len(apiKeys))
	for idx := range apiKeys {
		apiKey := apiKeys[idx]
		dtos[idx] = dtomaps.ToAccountApiKeyDto(&apiKey, nil)
	}

	return connect.NewResponse(&mgmtv1alpha1.GetAccountApiKeysResponse{
		ApiKeys: dtos,
	}), nil
}

func (s *Service) GetAccountApiKey(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccountApiKeyRequest],
) (*connect.Response[mgmtv1alpha1.GetAccountApiKeyResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("id", req.Msg.Id)
	_ = logger

	apiKeyUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}

	apiKey, err := s.db.Q.GetAccountApiKeyById(ctx, s.db.Db, apiKeyUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find api key")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(apiKey.AccountID))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetAccountApiKeyResponse{
		ApiKey: dtomaps.ToAccountApiKeyDto(&apiKey, nil),
	}), nil
}

func (s *Service) CreateAccountApiKey(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.CreateAccountApiKeyRequest],
) (*connect.Response[mgmtv1alpha1.CreateAccountApiKeyResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.CreateAccountApiKeyResponse{}), nil
}

func (s *Service) RegenerateAccountApiKey(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.RegenerateAccountApiKeyRequest],
) (*connect.Response[mgmtv1alpha1.RegenerateAccountApiKeyResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.RegenerateAccountApiKeyResponse{}), nil
}

func (s *Service) DeleteAccountApiKey(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.DeleteAccountApiKeyRequest],
) (*connect.Response[mgmtv1alpha1.DeleteAccountApiKeyResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("id", req.Msg.Id)
	_ = logger

	apiKeyUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	apiKey, err := s.db.Q.GetAccountApiKeyById(ctx, s.db.Db, apiKeyUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteAccountApiKeyResponse{}), nil
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(apiKey.AccountID))
	if err != nil {
		return nil, err
	}

	err = s.db.Q.RemoveAccountApiKey(ctx, s.db.Db, apiKeyUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.DeleteAccountApiKeyResponse{}), nil
}
