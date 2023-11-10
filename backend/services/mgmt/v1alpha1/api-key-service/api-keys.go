package v1alpha1_apikeyservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/nucleuscloud/neosync/backend/internal/utils"
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
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	expiresAt, err := nucleusdb.ToTimestamp(req.Msg.ExpiresAt.AsTime())
	if err != nil {
		return nil, err
	}

	clearKeyValue := getNewKeyValue()
	hashedKeyValue := utils.ToSha256(
		clearKeyValue,
	)

	newApiKey, err := s.db.Q.CreateAccountApiKey(ctx, s.db.Db, db_queries.CreateAccountApiKeyParams{
		KeyName:     req.Msg.Name,
		KeyValue:    hashedKeyValue,
		AccountID:   *accountUuid,
		ExpiresAt:   expiresAt,
		CreatedByID: *userUuid,
		UpdatedByID: *userUuid,
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mgmtv1alpha1.CreateAccountApiKeyResponse{
		ApiKey: dtomaps.ToAccountApiKeyDto(&newApiKey, &clearKeyValue),
	}), nil
}

func getNewKeyValue() string {
	return fmt.Sprintf("neo_at_v1_%s", uuid.New().String())
}

func (s *Service) RegenerateAccountApiKey(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.RegenerateAccountApiKeyRequest],
) (*connect.Response[mgmtv1alpha1.RegenerateAccountApiKeyResponse], error) {
	apiKeyUuid, err := nucleusdb.ToUuid(req.Msg.Id)
	if err != nil {
		return nil, err
	}

	apiKey, err := s.db.Q.GetAccountApiKeyById(ctx, s.db.Db, apiKeyUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("account api key not found")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(apiKey.AccountID))
	if err != nil {
		return nil, err
	}
	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}
	clearKeyValue := getNewKeyValue()
	hashedKeyValue := utils.ToSha256(
		clearKeyValue,
	)
	expiresAt, err := nucleusdb.ToTimestamp(req.Msg.ExpiresAt.AsTime())
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
