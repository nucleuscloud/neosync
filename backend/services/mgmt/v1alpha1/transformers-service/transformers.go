package v1alpha1_transformersservice

import (
	"context"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
)

type Transformation string

const (
	GenerateEmail          Transformation = "generate_email"
	GenerateRealisticEmail Transformation = "generate_realistic_email"
	TransformEmail         Transformation = "transform_email"
	GenerateBool           Transformation = "generate_bool"
	GenerateCardNumber     Transformation = "generate_card_number"
	GenerateCity           Transformation = "generate_city"
	GenerateE164Number     Transformation = "generate_e164_number"
	GenerateFirstName      Transformation = "generate_first_name"
	GenerateFloat          Transformation = "generate_float"
	GenerateFullAddress    Transformation = "generate_full_address"
	GenerateFullName       Transformation = "generate_full_name"
	GenerateGender         Transformation = "generate_gender"
	GenerateInt64Phone     Transformation = "generate_int64_phone"
	GenerateInt            Transformation = "generate_int"
	GenerateLastName       Transformation = "generate_last_name"
	GenerateShaHash256     Transformation = "generate_sha256hash"
	GenerateSSN            Transformation = "generate_ssn"
	GenerateState          Transformation = "generate_state"
	GenerateStreetAddress  Transformation = "generate_street_address"
	GenerateStringPhone    Transformation = "generate_string_phone"
	GenerateString         Transformation = "generate_string"
	GenerateUnixTimestamp  Transformation = "generate_unixtimestamp"
	GenerateUsername       Transformation = "generate_username"
	GenerateUtcTimestamp   Transformation = "generate_utctimestamp"
	GenerateUuid           Transformation = "generate_uuid"
	GenerateZipcode        Transformation = "generate_zipcode"
	TransformE164Phone     Transformation = "transform_e164_phone"
	TransformFirstName     Transformation = "transform_first_name"
	TransformFloat         Transformation = "transform_float"
	TransformFullName      Transformation = "transform_full_name"
	TransformIntPhone      Transformation = "transform_int_phone"
	TransformInt           Transformation = "transform_int"
	TransformLastName      Transformation = "transform_last_name"
	TransformPhone         Transformation = "transform_phone"
	TransformString        Transformation = "transform_string"
	Passthrough            Transformation = "passthrough"
	Null                   Transformation = "null"
	Invalid                Transformation = "invalid"
)

func (s *Service) GetSystemTransformers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetSystemTransformersRequest],
) (*connect.Response[mgmtv1alpha1.GetSystemTransformersResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.GetSystemTransformersResponse{
		Transformers: []*mgmtv1alpha1.Transformer{
			{Value: string(GenerateEmail), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
					GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{},
				},
			}},
			{Value: string(GenerateRealisticEmail), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateRealisticEmailConfig{
					GenerateRealisticEmailConfig: &mgmtv1alpha1.GenerateRealisticEmail{},
				},
			}},
			{Value: string(TransformEmail), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
					TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
						PreserveDomain: false,
						PreserveLength: false,
					},
				},
			}},
			{Value: string(GenerateBool), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
					GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
				},
			}},
			{Value: string(GenerateCardNumber), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{
					GenerateCardNumberConfig: &mgmtv1alpha1.GenerateCardNumber{
						ValidLuhn: true,
					},
				},
			}},
			{Value: string(GenerateCity), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
					GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
				},
			}},
			{Value: string(GenerateE164Number), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateE164NumberConfig{
					GenerateE164NumberConfig: &mgmtv1alpha1.GenerateE164Number{
						Length: 12,
					},
				},
			}},
			{Value: string(GenerateFirstName), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
					GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
				},
			}},
			{Value: string(GenerateFloat), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFloatConfig{
					GenerateFloatConfig: &mgmtv1alpha1.GenerateFloat{
						Sign:                "postiive",
						DigitsBeforeDecimal: 3,
						DigitsAfterDecimal:  3,
					},
				},
			}},
			{Value: string(GenerateFullAddress), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig{
					GenerateFullAddressConfig: &mgmtv1alpha1.GenerateFullAddress{},
				},
			}},
			{Value: string(GenerateFullName), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			}},
			{Value: string(GenerateGender), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateGenderConfig{
					GenerateGenderConfig: &mgmtv1alpha1.GenerateGender{
						Abbreviate: false,
					},
				},
			}},
			{Value: string(GenerateInt64Phone), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneConfig{
					GenerateInt64PhoneConfig: &mgmtv1alpha1.GenerateInt64Phone{},
				},
			}},
			{Value: string(GenerateInt), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateIntConfig{
					GenerateIntConfig: &mgmtv1alpha1.GenerateInt{
						Length: 4,
						Sign:   "positive",
					},
				},
			}},
			{Value: string(GenerateLastName), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig{
					GenerateLastNameConfig: &mgmtv1alpha1.GenerateLastName{},
				},
			}},
			{Value: string(GenerateShaHash256), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig{
					GenerateSha256HashConfig: &mgmtv1alpha1.GenerateSha256Hash{},
				},
			}},
			{Value: string(GenerateSSN), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
					GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
				},
			}},
			{Value: string(GenerateState), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStateConfig{
					GenerateStateConfig: &mgmtv1alpha1.GenerateState{},
				},
			}},
			{Value: string(GenerateStreetAddress), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig{
					GenerateStreetAddressConfig: &mgmtv1alpha1.GenerateStreetAddress{},
				},
			}},
			{Value: string(GenerateStringPhone), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringPhoneConfig{
					GenerateStringPhoneConfig: &mgmtv1alpha1.GenerateStringPhone{
						E164Format:     false,
						IncludeHyphens: false,
					},
				},
			}},
			{Value: string(GenerateString), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
					GenerateStringConfig: &mgmtv1alpha1.GenerateString{
						Length: 6,
					},
				},
			}},
			{Value: string(GenerateUnixTimestamp), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{
					GenerateUnixtimestampConfig: &mgmtv1alpha1.GenerateUnixTimestamp{},
				},
			}},
			{Value: string(GenerateUsername), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig{
					GenerateUsernameConfig: &mgmtv1alpha1.GenerateUsername{},
				},
			}},
			{Value: string(GenerateUtcTimestamp), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig{
					GenerateUtctimestampConfig: &mgmtv1alpha1.GenerateUtcTimestamp{},
				},
			}},
			{Value: string(GenerateUuid), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
					GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
						IncludeHyphens: true,
					},
				},
			}},
			{Value: string(GenerateZipcode), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig{
					GenerateZipcodeConfig: &mgmtv1alpha1.GenerateZipcode{},
				},
			}},
			{Value: string(TransformE164Phone), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformE164PhoneConfig{
					TransformE164PhoneConfig: &mgmtv1alpha1.TransformE164Phone{
						PreserveLength: false,
					},
				},
			}},
			{Value: string(TransformFirstName), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
					TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{
						PreserveLength: false,
					},
				},
			}},
			{Value: string(TransformFloat), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFloatConfig{
					TransformFloatConfig: &mgmtv1alpha1.TransformFloat{
						PreserveLength: false,
						PreserveSign:   true,
					},
				},
			}},
			{Value: string(TransformFullName), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
					TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{
						PreserveLength: false,
					},
				},
			}},
			{Value: string(TransformIntPhone), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformIntPhoneConfig{
					TransformIntPhoneConfig: &mgmtv1alpha1.TransformIntPhone{
						PreserveLength: false,
					},
				},
			}},
			{Value: string(TransformInt), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformIntConfig{
					TransformIntConfig: &mgmtv1alpha1.TransformInt{
						PreserveLength: false,
						PreserveSign:   true,
					},
				},
			}},
			{Value: string(TransformLastName), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformLastNameConfig{
					TransformLastNameConfig: &mgmtv1alpha1.TransformLastName{
						PreserveLength: false,
					},
				},
			}},
			{Value: string(TransformPhone), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformPhoneConfig{
					TransformPhoneConfig: &mgmtv1alpha1.TransformPhone{
						PreserveLength: false,
						IncludeHyphens: false,
					},
				},
			}},
			{Value: string(TransformString), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
					TransformStringConfig: &mgmtv1alpha1.TransformString{
						PreserveLength: false,
					},
				},
			}},
			{Value: string(Passthrough), Config: &mgmtv1alpha1.TransformerConfig{}},
			{Value: string(Null), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
					Nullconfig: &mgmtv1alpha1.Null{},
				},
			}},
		},
	}), nil
}

func (s *Service) GetCustomTransformers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetCustomTransformersRequest],
) (*connect.Response[mgmtv1alpha1.GetCustomTransformersResponse], error) {

	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	transformers, err := s.db.Q.GetCustomTransformersByAccount(ctx, s.db.Db, *accountUuid)
	if err != nil {
		return nil, err
	}

	dtoTransformers := []*mgmtv1alpha1.CustomTransformer{}
	for idx := range transformers {
		transformer := transformers[idx]
		dtoTransformers = append(dtoTransformers, dtomaps.ToCustomTransformerDto(&transformer))
	}

	return connect.NewResponse(&mgmtv1alpha1.GetCustomTransformersResponse{
		Transformers: dtoTransformers,
	}), nil
}

func (s *Service) GetCustomTransformerById(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetCustomTransformerByIdRequest],
) (*connect.Response[mgmtv1alpha1.GetCustomTransformerByIdResponse], error) {

	tId, err := nucleusdb.ToUuid(req.Msg.TransformerId)
	if err != nil {
		return nil, err
	}

	transformer, err := s.db.Q.GetCustomTransformerById(ctx, s.db.Db, tId)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.GetCustomTransformerByIdResponse{}), nil
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(transformer.AccountID))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.GetCustomTransformerByIdResponse{
		Transformer: dtomaps.ToCustomTransformerDto(&transformer),
	}), nil
}

func (s *Service) CreateCustomTransformer(ctx context.Context, req *connect.Request[mgmtv1alpha1.CreateCustomTransformerRequest]) (*connect.Response[mgmtv1alpha1.CreateCustomTransformerResponse], error) {

	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	customTransformer := &db_queries.CreateCustomTransformerParams{
		AccountID:         *accountUuid,
		Name:              req.Msg.Name,
		Description:       req.Msg.Description,
		TransformerConfig: &pg_models.TransformerConfigs{},
		Type:              req.Msg.Type,
		Source:            req.Msg.Source,
		CreatedByID:       *userUuid,
		UpdatedByID:       *userUuid,
	}

	err = customTransformer.TransformerConfig.FromTransformerConfigDto(req.Msg.TransformerConfig)
	if err != nil {
		return nil, err
	}

	ct, err := s.db.Q.CreateCustomTransformer(ctx, s.db.Db, *customTransformer)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.CreateCustomTransformerResponse{
		Transformer: dtomaps.ToCustomTransformerDto(&ct),
	}), nil

}

func (s *Service) DeleteCustomTransformer(ctx context.Context, req *connect.Request[mgmtv1alpha1.DeleteCustomTransformerRequest]) (*connect.Response[mgmtv1alpha1.DeleteCustomTransformerResponse], error) {

	logger := logger_interceptor.GetLoggerFromContextOrDefault(ctx)
	logger = logger.With("transformer", req.Msg.TransformerId)

	tId, err := nucleusdb.ToUuid(req.Msg.TransformerId)
	if err != nil {
		return nil, err
	}

	transformer, err := s.db.Q.GetCustomTransformerById(ctx, s.db.Db, tId)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return connect.NewResponse(&mgmtv1alpha1.DeleteCustomTransformerResponse{}), nil
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(transformer.AccountID))
	if err != nil {
		return nil, err
	}

	err = s.db.Q.DeleteCustomTransformerById(ctx, s.db.Db, transformer.ID)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		logger.Info("destination not found")
	}

	return connect.NewResponse(&mgmtv1alpha1.DeleteCustomTransformerResponse{}), nil

}

func (s *Service) UpdateCustomTransformer(ctx context.Context, req *connect.Request[mgmtv1alpha1.UpdateCustomTransformerRequest]) (*connect.Response[mgmtv1alpha1.UpdateCustomTransformerResponse], error) {

	tUuid, err := nucleusdb.ToUuid(req.Msg.TransformerId)
	if err != nil {
		return nil, err
	}
	transformer, err := s.db.Q.GetCustomTransformerById(ctx, s.db.Db, tUuid)
	if err != nil && !nucleusdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && nucleusdb.IsNoRows(err) {
		return nil, nucleuserrors.NewNotFound("unable to find connection by id")
	}

	_, err = s.verifyUserInAccount(ctx, nucleusdb.UUIDString(transformer.AccountID))
	if err != nil {
		return nil, err
	}

	userUuid, err := s.getUserUuid(ctx)
	if err != nil {
		return nil, err
	}

	customTransformer := &db_queries.UpdateCustomTransformerParams{
		Name:              req.Msg.Name,
		Description:       req.Msg.Description,
		TransformerConfig: &pg_models.TransformerConfigs{},
		UpdatedByID:       *userUuid,
		ID:                tUuid,
	}

	err = customTransformer.TransformerConfig.FromTransformerConfigDto(req.Msg.TransformerConfig)
	if err != nil {
		return nil, err
	}

	t, err := s.db.Q.UpdateCustomTransformer(ctx, s.db.Db, *customTransformer)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.UpdateCustomTransformerResponse{
		Transformer: dtomaps.ToCustomTransformerDto(&t),
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
