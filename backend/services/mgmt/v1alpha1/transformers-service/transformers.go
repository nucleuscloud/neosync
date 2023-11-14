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
	Invalid              Transformation = "invalid"
	Passthrough          Transformation = "passthrough"
	Uuid                 Transformation = "uuid"
	FirstName            Transformation = "first_name"
	LastName             Transformation = "last_name"
	FullName             Transformation = "full_name"
	PhoneNumber          Transformation = "phone_number"
	IntPhoneNumber       Transformation = "int_phone_number"
	Email                Transformation = "email"
	Null                 Transformation = "null"
	RandomString         Transformation = "random_string"
	RandomBool           Transformation = "random_bool"
	RandomInt            Transformation = "random_int"
	RandomFloat          Transformation = "random_float"
	Gender               Transformation = "gender"
	UTCTimestamp         Transformation = "utc_timestamp"
	UnixTimestamp        Transformation = "unix_timestamp"
	StreetAddress        Transformation = "street_address"
	City                 Transformation = "city"
	Zipcode              Transformation = "zipcode"
	State                Transformation = "state"
	FullAddress          Transformation = "full_address"
	CreditCard           Transformation = "credit_card"
	SHA256               Transformation = "sha256_hash"
	SocialSecurityNumber Transformation = "social_security_number"
)

func (s *Service) GetSystemTransformers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetSystemTransformersRequest],
) (*connect.Response[mgmtv1alpha1.GetSystemTransformersResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.GetSystemTransformersResponse{
		Transformers: []*mgmtv1alpha1.Transformer{
			{Value: string(Passthrough), Config: &mgmtv1alpha1.TransformerConfig{}},
			{Value: string(Uuid), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_UuidConfig{
					UuidConfig: &mgmtv1alpha1.Uuid{
						IncludeHyphen: true,
					},
				},
			},
			},
			{Value: string(FirstName), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_FirstNameConfig{
					FirstNameConfig: &mgmtv1alpha1.FirstName{
						PreserveLength: false,
					},
				},
			}},
			{Value: string(LastName), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_LastNameConfig{
					LastNameConfig: &mgmtv1alpha1.LastName{
						PreserveLength: false,
					},
				},
			}},
			{Value: string(FullName), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_FullNameConfig{
					FullNameConfig: &mgmtv1alpha1.FullName{
						PreserveLength: false,
					},
				},
			}},
			{Value: string(PhoneNumber), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_PhoneNumberConfig{
					PhoneNumberConfig: &mgmtv1alpha1.PhoneNumber{
						PreserveLength: false,
						E164Format:     false,
						IncludeHyphens: false,
					},
				},
			}},
			{Value: string(IntPhoneNumber), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_IntPhoneNumberConfig{
					IntPhoneNumberConfig: &mgmtv1alpha1.IntPhoneNumber{
						PreserveLength: false,
					},
				},
			}},
			{Value: string(Null), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_NullConfig{
					NullConfig: &mgmtv1alpha1.Null{},
				},
			}},
			{
				Value: string(Email),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_EmailConfig{
						EmailConfig: &mgmtv1alpha1.EmailConfig{
							PreserveDomain: false,
							PreserveLength: false,
						},
					},
				}},
			{
				Value: string(RandomString),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_RandomStringConfig{
						RandomStringConfig: &mgmtv1alpha1.RandomString{
							PreserveLength: false,
							StrLength:      0,
						},
					},
				}},
			{Value: string(RandomBool), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_RandomBoolConfig{
					RandomBoolConfig: &mgmtv1alpha1.RandomBool{},
				},
			}},
			{
				Value: string(RandomInt),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_RandomIntConfig{
						RandomIntConfig: &mgmtv1alpha1.RandomInt{
							PreserveLength: false,
							IntLength:      0,
						},
					},
				}},
			{
				Value: string(RandomFloat),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_RandomFloatConfig{
						RandomFloatConfig: &mgmtv1alpha1.RandomFloat{
							PreserveLength:      false,
							DigitsBeforeDecimal: 2,
							DigitsAfterDecimal:  3,
						},
					},
				}},
			{
				Value: string(Gender),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenderConfig{
						GenderConfig: &mgmtv1alpha1.Gender{
							Abbreviate: false,
						},
					},
				}},
			{Value: string(UTCTimestamp), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_UtcTimestampConfig{
					UtcTimestampConfig: &mgmtv1alpha1.UTCTimestamp{},
				},
			}},
			{Value: string(UnixTimestamp), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_UnixTimestampConfig{
					UnixTimestampConfig: &mgmtv1alpha1.UnixTimestamp{},
				},
			}},
			{Value: string(StreetAddress), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_StreetAddressConfig{
					StreetAddressConfig: &mgmtv1alpha1.StreetAddress{},
				},
			}},
			{Value: string(City), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_CityConfig{
					CityConfig: &mgmtv1alpha1.City{},
				},
			}},
			{Value: string(Zipcode), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_ZipcodeConfig{
					ZipcodeConfig: &mgmtv1alpha1.Zipcode{},
				},
			}},
			{Value: string(State), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_StateConfig{
					StateConfig: &mgmtv1alpha1.State{},
				},
			}},
			{Value: string(FullAddress), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_FullAddressConfig{
					FullAddressConfig: &mgmtv1alpha1.FullAddress{},
				},
			}},
			{Value: string(CreditCard), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_CreditCardConfig{
					CreditCardConfig: &mgmtv1alpha1.CreditCard{
						ValidLuhn: true,
					},
				},
			}},
			{Value: string(SHA256), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_Sha256HashConfig{
					Sha256HashConfig: &mgmtv1alpha1.SHA256Hash{},
				},
			}},
			{Value: string(SocialSecurityNumber), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_SsnConfig{
					SsnConfig: &mgmtv1alpha1.SocialSecurityNumber{},
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
