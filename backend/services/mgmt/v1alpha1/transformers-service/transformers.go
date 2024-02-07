package v1alpha1_transformersservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/dop251/goja"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
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
		dtoTransformers = append(dtoTransformers, dtomaps.ToUserDefinedTransformerDto(&transformer))
	}

	return connect.NewResponse(&mgmtv1alpha1.GetUserDefinedTransformersResponse{
		Transformers: dtoTransformers,
	}), nil
}

func (s *Service) GetUserDefinedTransformerById(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetUserDefinedTransformerByIdRequest],
) (*connect.Response[mgmtv1alpha1.GetUserDefinedTransformerByIdResponse], error) {
	tId, err := nucleusdb.ToUuid(req.Msg.TransformerId)
	if err != nil {
		return nil, err
	}

	transformer, err := s.db.Q.GetUserDefinedTransformerById(ctx, s.db.Db, tId)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find transformer by id")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(transformer.AccountID))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetUserDefinedTransformerByIdResponse{
		Transformer: dtomaps.ToUserDefinedTransformerDto(&transformer),
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
		TransformerConfig: &pg_models.TransformerConfigs{},
		Type:              req.Msg.Type,
		Source:            req.Msg.Source,
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

	return connect.NewResponse(&mgmtv1alpha1.CreateUserDefinedTransformerResponse{
		Transformer: dtomaps.ToUserDefinedTransformerDto(&ct),
	}), nil

}

func (s *Service) DeleteUserDefinedTransformer(ctx context.Context, req *connect.Request[mgmtv1alpha1.DeleteUserDefinedTransformerRequest]) (*connect.Response[mgmtv1alpha1.DeleteUserDefinedTransformerResponse], error) {
	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("transformerId", req.Msg.TransformerId)

	tId, err := nucleusdb.ToUuid(req.Msg.TransformerId)
	if err != nil {
		return nil, err
	}

	transformer, err := s.db.Q.GetUserDefinedTransformerById(ctx, s.db.Db, tId)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteUserDefinedTransformerResponse{}), nil
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(transformer.AccountID))
	if err != nil {
		return nil, err
	}

	err = s.db.Q.DeleteUserDefinedTransformerById(ctx, s.db.Db, transformer.ID)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		logger.Info("transformer not found or has already been removed")
	}

	return connect.NewResponse(&mgmtv1alpha1.DeleteUserDefinedTransformerResponse{}), nil
}

func (s *Service) UpdateUserDefinedTransformer(ctx context.Context, req *connect.Request[mgmtv1alpha1.UpdateUserDefinedTransformerRequest]) (*connect.Response[mgmtv1alpha1.UpdateUserDefinedTransformerResponse], error) {
	tUuid, err := nucleusdb.ToUuid(req.Msg.TransformerId)
	if err != nil {
		return nil, err
	}
	transformer, err := s.db.Q.GetUserDefinedTransformerById(ctx, s.db.Db, tUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find transformer by id")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(transformer.AccountID))
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
		TransformerConfig: &pg_models.TransformerConfigs{},
		UpdatedByID:       *userUuid,
		ID:                tUuid,
	}
	err = updateParams.TransformerConfig.FromTransformerConfigDto(req.Msg.TransformerConfig)
	if err != nil {
		return nil, err
	}

	updatedTransformer, err := s.db.Q.UpdateUserDefinedTransformer(ctx, s.db.Db, *updateParams)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.UpdateUserDefinedTransformerResponse{
		Transformer: dtomaps.ToUserDefinedTransformerDto(&updatedTransformer),
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
	_, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	js := constructJavascriptCode(req.Msg.Code)

	_, err = goja.Compile("test", js, true)
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
