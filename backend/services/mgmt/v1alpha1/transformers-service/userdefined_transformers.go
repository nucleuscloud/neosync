package v1alpha1_transformersservice

import (
	"context"
	"fmt"
	"regexp"

	"connectrpc.com/connect"
	"github.com/dop251/goja"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
)

func (s *Service) GetUserDefinedTransformers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetUserDefinedTransformersRequest],
) (*connect.Response[mgmtv1alpha1.GetUserDefinedTransformersResponse], error) {
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	transformers, err := s.db.Q.GetUserDefinedTransformersByAccount(ctx, s.db.Db, *accountUuid)
	if err != nil {
		return nil, err
	}

	dtoTransformers := []*mgmtv1alpha1.UserDefinedTransformer{}
	for idx := range transformers {
		transformer := transformers[idx]
		dto, err := dtomaps.ToUserDefinedTransformerDto(&transformer, baseSystemTransformerSourceMap)
		if err != nil {
			return nil, fmt.Errorf("failed to map user defined transformer %s with source %d", neosyncdb.UUIDString(transformer.ID), transformer.Source)
		}
		dtoTransformers = append(dtoTransformers, dto)
	}

	return connect.NewResponse(&mgmtv1alpha1.GetUserDefinedTransformersResponse{
		Transformers: dtoTransformers,
	}), nil
}

func (s *Service) GetUserDefinedTransformerById(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetUserDefinedTransformerByIdRequest],
) (*connect.Response[mgmtv1alpha1.GetUserDefinedTransformerByIdResponse], error) {
	tId, err := neosyncdb.ToUuid(req.Msg.TransformerId)
	if err != nil {
		return nil, err
	}

	transformer, err := s.db.Q.GetUserDefinedTransformerById(ctx, s.db.Db, tId)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find transformer by id")
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(transformer.AccountID))
	if err != nil {
		return nil, err
	}

	dto, err := dtomaps.ToUserDefinedTransformerDto(&transformer, baseSystemTransformerSourceMap)
	if err != nil {
		return nil, fmt.Errorf("failed to map user defined transformer %s with source %d", neosyncdb.UUIDString(transformer.ID), transformer.Source)
	}

	return connect.NewResponse(&mgmtv1alpha1.GetUserDefinedTransformerByIdResponse{
		Transformer: dto,
	}), nil
}

func (s *Service) CreateUserDefinedTransformer(ctx context.Context, req *connect.Request[mgmtv1alpha1.CreateUserDefinedTransformerRequest]) (*connect.Response[mgmtv1alpha1.CreateUserDefinedTransformerResponse], error) {
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	UserDefinedTransformer := &db_queries.CreateUserDefinedTransformerParams{
		AccountID:         *accountUuid,
		Name:              req.Msg.Name,
		Description:       req.Msg.Description,
		TransformerConfig: &pg_models.TransformerConfig{},
		Source:            int32(req.Msg.Source),
		CreatedByID:       *userUuid,
		UpdatedByID:       *userUuid,
	}

	err = UserDefinedTransformer.TransformerConfig.FromTransformerConfigDto(req.Msg.TransformerConfig)
	if err != nil {
		return nil, err
	}

	ct, err := s.db.Q.CreateUserDefinedTransformer(ctx, s.db.Db, *UserDefinedTransformer)
	if err != nil {
		return nil, err
	}

	dto, err := dtomaps.ToUserDefinedTransformerDto(&ct, baseSystemTransformerSourceMap)
	if err != nil {
		return nil, fmt.Errorf("failed to map user defined transformer %s with source %d", neosyncdb.UUIDString(ct.ID), ct.Source)
	}

	return connect.NewResponse(&mgmtv1alpha1.CreateUserDefinedTransformerResponse{
		Transformer: dto,
	}), nil
}

func (s *Service) DeleteUserDefinedTransformer(ctx context.Context, req *connect.Request[mgmtv1alpha1.DeleteUserDefinedTransformerRequest]) (*connect.Response[mgmtv1alpha1.DeleteUserDefinedTransformerResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("transformerId", req.Msg.TransformerId)

	tId, err := neosyncdb.ToUuid(req.Msg.TransformerId)
	if err != nil {
		return nil, err
	}

	transformer, err := s.db.Q.GetUserDefinedTransformerById(ctx, s.db.Db, tId)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteUserDefinedTransformerResponse{}), nil
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(transformer.AccountID))
	if err != nil {
		return nil, err
	}

	err = s.db.Q.DeleteUserDefinedTransformerById(ctx, s.db.Db, transformer.ID)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		logger.Info("transformer not found or has already been removed")
	}

	return connect.NewResponse(&mgmtv1alpha1.DeleteUserDefinedTransformerResponse{}), nil
}

func (s *Service) UpdateUserDefinedTransformer(ctx context.Context, req *connect.Request[mgmtv1alpha1.UpdateUserDefinedTransformerRequest]) (*connect.Response[mgmtv1alpha1.UpdateUserDefinedTransformerResponse], error) {
	tUuid, err := neosyncdb.ToUuid(req.Msg.TransformerId)
	if err != nil {
		return nil, err
	}
	transformer, err := s.db.Q.GetUserDefinedTransformerById(ctx, s.db.Db, tUuid)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find transformer by id")
	}

	_, err = s.verifyUserInAccount(ctx, neosyncdb.UUIDString(transformer.AccountID))
	if err != nil {
		return nil, err
	}

	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	updateParams := &db_queries.UpdateUserDefinedTransformerParams{
		Name:              req.Msg.Name,
		Description:       req.Msg.Description,
		TransformerConfig: &pg_models.TransformerConfig{},
		UpdatedByID:       *userUuid,
		ID:                tUuid,
	}
	// todo: must verify that this updated config is valid for the configured source
	err = updateParams.TransformerConfig.FromTransformerConfigDto(req.Msg.TransformerConfig)
	if err != nil {
		return nil, err
	}

	updatedTransformer, err := s.db.Q.UpdateUserDefinedTransformer(ctx, s.db.Db, *updateParams)
	if err != nil {
		return nil, err
	}

	dto, err := dtomaps.ToUserDefinedTransformerDto(&updatedTransformer, baseSystemTransformerSourceMap)
	if err != nil {
		return nil, fmt.Errorf("failed to map user defined transformer %s with source %d", neosyncdb.UUIDString(updatedTransformer.ID), updatedTransformer.Source)
	}

	return connect.NewResponse(&mgmtv1alpha1.UpdateUserDefinedTransformerResponse{
		Transformer: dto,
	}), err
}

func (s *Service) IsTransformerNameAvailable(ctx context.Context, req *connect.Request[mgmtv1alpha1.IsTransformerNameAvailableRequest]) (*connect.Response[mgmtv1alpha1.IsTransformerNameAvailableResponse], error) {
	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	count, err := s.db.Q.IsTransformerNameAvailable(ctx, s.db.Db, db_queries.IsTransformerNameAvailableParams{
		AccountId:       *accountUuid,
		TransformerName: req.Msg.TransformerName,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.IsTransformerNameAvailableResponse{
		IsAvailable: count == 0,
	}), nil
}

// use the goja library to validate that the javascript can compile and theoretically run
func (s *Service) ValidateUserJavascriptCode(ctx context.Context, req *connect.Request[mgmtv1alpha1.ValidateUserJavascriptCodeRequest]) (*connect.Response[mgmtv1alpha1.ValidateUserJavascriptCodeResponse], error) {
	js := constructJavascriptCode(req.Msg.GetCode())

	_, err := goja.Compile("test", js, true)
	if err != nil {
		return connect.NewResponse(&mgmtv1alpha1.ValidateUserJavascriptCodeResponse{
			Valid: false,
		}), nil
	}

	return connect.NewResponse(&mgmtv1alpha1.ValidateUserJavascriptCodeResponse{
		Valid: true,
	}), nil
}

func constructJavascriptCode(jsCode string) string {
	if jsCode != "" {
		return fmt.Sprintf(`(()=>{
			function fn1(value){
				%s
				}})();`, jsCode)
	} else {
		return ""
	}
}

func (s *Service) ValidateUserRegexCode(ctx context.Context, req *connect.Request[mgmtv1alpha1.ValidateUserRegexCodeRequest]) (*connect.Response[mgmtv1alpha1.ValidateUserRegexCodeResponse], error) {
	_, err := regexp.Compile(req.Msg.GetUserProvidedRegex())
	// todo: should return error message here and surface to user
	return connect.NewResponse(&mgmtv1alpha1.ValidateUserRegexCodeResponse{
		Valid: err == nil,
	}), nil
}
