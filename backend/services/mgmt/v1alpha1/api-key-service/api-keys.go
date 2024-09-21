package v1alpha1_apikeyservice

import (
	"context"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/apikey"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	pkg_utils "github.com/nucleuscloud/neosync/backend/pkg/utils"
)

func (s *Service) GetAccountApiKeys(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetAccountApiKeysRequest],
) (*connect.Response[mgmtv1alpha1.GetAccountApiKeysResponse], error) {
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	apiKeys, err := s.db.Q.GetAccountApiKeys(ctx, s.db.Db, *accountUuid)
	if err != nil {
		return nil, err
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
	apiKeyUuid, err := neosyncdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}

	apiKey, err := s.db.Q.GetAccountApiKeyById(ctx, s.db.Db, apiKeyUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find api key")
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(apiKey.AccountID))
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
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	expiresAt, err := neosyncdb.ToTimestamp(req.Msg.ExpiresAt.AsTime())
	if err != nil {
		return nil, err
	}

	clearKeyValue := apikey.NewV1AccountKey()
	hashedKeyValue := pkg_utils.ToSha256(
		clearKeyValue,
	)

	newApiKey, err := s.db.CreateAccountApikey(ctx, &neosyncdb.CreateAccountApiKeyRequest{
		KeyName:           req.Msg.Name,
		KeyValue:          hashedKeyValue,
		AccountUuid:       *accountUuid,
		CreatedByUserUuid: *userUuid,
		ExpiresAt:         expiresAt,
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.CreateAccountApiKeyResponse{
		ApiKey: dtomaps.ToAccountApiKeyDto(newApiKey, &clearKeyValue),
	}), nil
}

func (s *Service) RegenerateAccountApiKey(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.RegenerateAccountApiKeyRequest],
) (*connect.Response[mgmtv1alpha1.RegenerateAccountApiKeyResponse], error) {
	apiKeyUuid, err := neosyncdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}

	apiKey, err := s.db.Q.GetAccountApiKeyById(ctx, s.db.Db, apiKeyUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("account api key not found")
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(apiKey.AccountID))
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}
	clearKeyValue := apikey.NewV1AccountKey()
	hashedKeyValue := pkg_utils.ToSha256(
		clearKeyValue,
	)
	expiresAt, err := neosyncdb.ToTimestamp(req.Msg.ExpiresAt.AsTime())
	if err != nil {
		return nil, err
	}
	updatedApiKey, err := s.db.Q.UpdateAccountApiKeyValue(ctx, s.db.Db, db_queries.UpdateAccountApiKeyValueParams{
		KeyValue:    hashedKeyValue,
		ExpiresAt:   expiresAt,
		UpdatedByID: *userUuid,
		ID:          apiKeyUuid,
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.RegenerateAccountApiKeyResponse{
		ApiKey: dtomaps.ToAccountApiKeyDto(&updatedApiKey, &clearKeyValue),
	}), nil
}

func (s *Service) DeleteAccountApiKey(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.DeleteAccountApiKeyRequest],
) (*connect.Response[mgmtv1alpha1.DeleteAccountApiKeyResponse], error) {
	apiKeyUuid, err := neosyncdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	apiKey, err := s.db.Q.GetAccountApiKeyById(ctx, s.db.Db, apiKeyUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteAccountApiKeyResponse{}), nil
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(apiKey.AccountID))
	if err != nil {
		return nil, err
	}

	err = s.db.Q.RemoveAccountApiKey(ctx, s.db.Db, apiKeyUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.DeleteAccountApiKeyResponse{}), nil
}
