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
	GenerateEmail             Transformation = "generate_email"
	TransformEmail            Transformation = "transform_email"
	GenerateBool              Transformation = "generate_bool"
	GenerateCardNumber        Transformation = "generate_card_number"
	GenerateCity              Transformation = "generate_city"
	GenerateDefault           Transformation = "generate_default"
	GenerateE164PhoneNumber   Transformation = "generate_e164_phone_number"
	GenerateFirstName         Transformation = "generate_first_name"
	GenerateFloat64           Transformation = "generate_float64"
	GenerateFullAddress       Transformation = "generate_full_address"
	GenerateFullName          Transformation = "generate_full_name"
	GenerateGender            Transformation = "generate_gender"
	GenerateInt64PhoneNumber  Transformation = "generate_int64_phone_number"
	GenerateInt64             Transformation = "generate_int64"
	GenerateLastName          Transformation = "generate_last_name"
	GenerateShaHash256        Transformation = "generate_sha256hash"
	GenerateSSN               Transformation = "generate_ssn"
	GenerateState             Transformation = "generate_state"
	GenerateStreetAddress     Transformation = "generate_street_address"
	GenerateStringPhoneNumber Transformation = "generate_string_phone_number"
	GenerateString            Transformation = "generate_string"
	GenerateUnixTimestamp     Transformation = "generate_unixtimestamp"
	GenerateUsername          Transformation = "generate_username"
	GenerateUtcTimestamp      Transformation = "generate_utctimestamp"
	GenerateUuid              Transformation = "generate_uuid"
	GenerateZipcode           Transformation = "generate_zipcode"
	TransformE164PhoneNumber  Transformation = "transform_e164_phone_number"
	TransformFirstName        Transformation = "transform_first_name"
	TransformFloat64          Transformation = "transform_float64"
	TransformFullName         Transformation = "transform_full_name"
	TransformInt64PhoneNumber Transformation = "transform_int64_phone_number"
	TransformInt64            Transformation = "transform_int64"
	TransformLastName         Transformation = "transform_last_name"
	TransformPhoneNumber      Transformation = "transform_phone_number"
	TransformString           Transformation = "transform_string"
	Passthrough               Transformation = "passthrough"
	Null                      Transformation = "null"
	Invalid                   Transformation = "invalid"
)

func (s *Service) GetSystemTransformers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetSystemTransformersRequest],
) (*connect.Response[mgmtv1alpha1.GetSystemTransformersResponse], error) {

	return connect.NewResponse(&mgmtv1alpha1.GetSystemTransformersResponse{
		Transformers: []*mgmtv1alpha1.SystemTransformer{
			{

				Name:        "Generate Email",
				Description: "Generates a new randomized email address.",
				DataType:    "string",
				Source:      string(GenerateEmail),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
						GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{},
					},
				},
			},
			{
				Name:        "Transform Email",
				Description: "Transforms an existing email address.",
				DataType:    "string",
				Source:      string(TransformEmail),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
						TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
							PreserveDomain: false,
							PreserveLength: false,
						},
					},
				},
			},
			{
				Name:        "Generate Boolean",
				Description: "Generates a boolean value at random.",
				DataType:    "boolean",
				Source:      string(GenerateBool),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
						GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
					},
				},
			},
			{
				Name:        "Generate Card Number",
				Description: "Generates a card number.",
				DataType:    "int64",
				Source:      string(GenerateCardNumber),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{
						GenerateCardNumberConfig: &mgmtv1alpha1.GenerateCardNumber{
							ValidLuhn: true,
						},
					},
				},
			},
			{
				Name:        "Generate City",
				Description: "Randomly selects a city from a list of predfined US cities.",
				DataType:    "string",
				Source:      string(GenerateCity),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
						GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
					},
				},
			},
			{
				Name:        "Generate Default",
				Description: "Defers to the database column default",
				DataType:    "string",
				Source:      string(GenerateDefault),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{
						GenerateDefaultConfig: &mgmtv1alpha1.GenerateDefault{},
					},
				},
			},
			{
				Name:        "Generate E164 Phone Number",
				Description: "Generates a Generate phone number in e164 format.",
				DataType:    "string",
				Source:      string(GenerateE164PhoneNumber),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateE164PhoneNumberConfig{
						GenerateE164PhoneNumberConfig: &mgmtv1alpha1.GenerateE164PhoneNumber{
							Min: 9,
							Max: 15,
						},
					},
				},
			},
			{
				Name:        "Generate First Name",
				Description: "Generates a random first name. ",
				DataType:    "string",
				Source:      string(GenerateFirstName),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
						GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
					},
				},
			},
			{
				Name:        "Generate Float64",
				Description: "Generates a random float64 value.",
				DataType:    "float64",
				Source:      string(GenerateFloat64),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{
						GenerateFloat64Config: &mgmtv1alpha1.GenerateFloat64{
							RandomizeSign: false,
							Min:           1.00,
							Max:           100.00,
						},
					},
				},
			},
			{
				Name:        "Generate Full Address",
				Description: "Randomly generates a street address in the format: {street_num} {street_addresss} {street_descriptor} {city}, {state} {zipcode}. For example, 123 Main Street Boston, Massachusetts 02169.",
				DataType:    "string",
				Source:      string(GenerateFullAddress),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig{
						GenerateFullAddressConfig: &mgmtv1alpha1.GenerateFullAddress{},
					},
				},
			},
			{
				Name:        "Generate Full Name",
				Description: "Generates a new full name consisting of a first and last name",
				DataType:    "string",
				Source:      string(GenerateFullName),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
						GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
					},
				},
			},
			{
				Name:        "Generate Gender",
				Description: "Randomly generates one of the following genders: female, male, undefined, nonbinary.",
				DataType:    "string",
				Source:      string(GenerateGender),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateGenderConfig{
						GenerateGenderConfig: &mgmtv1alpha1.GenerateGender{
							Abbreviate: false,
						},
					},
				},
			},
			{
				Name:        "Generate Int64 Phone Number",
				Description: "Generates a new phone number with a default length of 10.",
				DataType:    "int64",
				Source:      string(GenerateInt64PhoneNumber),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneNumberConfig{
						GenerateInt64PhoneNumberConfig: &mgmtv1alpha1.GenerateInt64PhoneNumber{},
					},
				},
			},
			{
				Name:        "Generate Random Int64",
				Description: "Generates a random int64 value.", DataType: "int64",
				Source: string(GenerateInt64),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
						GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
							RandomizeSign: false,
							Min:           1,
							Max:           4,
						},
					},
				},
			},
			{
				Name:        "Generate Last Name",
				Description: "Generates a random last name.", DataType: "int64",
				Source: string(GenerateLastName),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig{
						GenerateLastNameConfig: &mgmtv1alpha1.GenerateLastName{},
					},
				},
			},
			{
				Name:        "Generate SHA256 Hash",
				Description: "SHA256 hashes a randomly generated value.",
				DataType:    "string",
				Source:      string(GenerateShaHash256),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig{
						GenerateSha256HashConfig: &mgmtv1alpha1.GenerateSha256Hash{},
					},
				},
			},
			{
				Name:        "Generate SSN",
				Description: "Generates a completely random social security numbers including the hyphens in the format <xxx-xx-xxxx>",
				DataType:    "string",
				Source:      string(GenerateSSN),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
						GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
					},
				},
			},
			{
				Name:        "Generate State",
				Description: "Randomly selects a US state and returns the two-character state code.",
				DataType:    "string",
				Source:      string(GenerateState),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateStateConfig{
						GenerateStateConfig: &mgmtv1alpha1.GenerateState{},
					},
				},
			},
			{
				Name:        "Generate Street Address",
				Description: "Randomly generates a street address in the format: {street_num} {street_addresss} {street_descriptor}. For example, 123 Main Street.",
				DataType:    "string",
				Source:      string(GenerateStreetAddress),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig{
						GenerateStreetAddressConfig: &mgmtv1alpha1.GenerateStreetAddress{},
					},
				},
			},
			{
				Name:        "Generate String Phone Number",
				Description: "Generates a Generate phone number and returns it as a string.",
				DataType:    "string",
				Source:      string(GenerateStringPhoneNumber),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig{
						GenerateStringPhoneNumberConfig: &mgmtv1alpha1.GenerateStringPhoneNumber{
							IncludeHyphens: false,
						},
					},
				},
			},
			{
				Name:        "Generate Random String",
				Description: "Creates a randomly ordered alphanumeric string with a default length of 10 unless the String Length parameter are defined.",
				DataType:    "string",
				Source:      string(GenerateString),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
						GenerateStringConfig: &mgmtv1alpha1.GenerateString{
							Min: 2,
							Max: 7,
						},
					},
				},
			},
			{
				Name:        "Generate Unix Timestamp",
				Description: "Randomly generates a Unix timestamp",
				DataType:    "int64",
				Source:      string(GenerateUnixTimestamp),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{
						GenerateUnixtimestampConfig: &mgmtv1alpha1.GenerateUnixTimestamp{},
					},
				},
			},
			{
				Name:        "Generate Username",
				Description: "Randomly generates a username in the format<first_initial><last_name>.",
				DataType:    "string",
				Source:      string(GenerateUsername),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig{
						GenerateUsernameConfig: &mgmtv1alpha1.GenerateUsername{},
					},
				},
			},
			{
				Name:        "Generate UTC Timestamp",
				Description: "Randomly generates a UTC timestamp.",
				DataType:    "time",
				Source:      string(GenerateUtcTimestamp),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig{
						GenerateUtctimestampConfig: &mgmtv1alpha1.GenerateUtcTimestamp{},
					},
				},
			},
			{
				Name:        "Generate UUID",
				Description: "Generates a new UUIDv4 id.",
				DataType:    "uuid",
				Source:      string(GenerateUuid),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
						GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
							IncludeHyphens: true,
						},
					},
				},
			},
			{
				Name:        "Generate Zipcode",
				Description: "Randomly selects a zip code from a list of predefined US zipcodes.",
				DataType:    "string",
				Source:      string(GenerateZipcode),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig{
						GenerateZipcodeConfig: &mgmtv1alpha1.GenerateZipcode{},
					},
				},
			},
			{
				Name:        "Transform E164 Phone Number",
				Description: "Transforms an existing E164 formatted phone number.",
				DataType:    "string",
				Source:      string(TransformE164PhoneNumber),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig{
						TransformE164PhoneNumberConfig: &mgmtv1alpha1.TransformE164PhoneNumber{
							PreserveLength: false,
						},
					},
				},
			},
			{
				Name:        "Transform First Name",
				Description: "Transforms an existing first name",
				DataType:    "string",
				Source:      string(TransformFirstName),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
						TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{
							PreserveLength: false,
						},
					},
				},
			},
			{
				Name:        "Transform Float64",
				Description: "Transforms an existing float value.",
				DataType:    "float64",
				Source:      string(TransformFloat64),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformFloat64Config{
						TransformFloat64Config: &mgmtv1alpha1.TransformFloat64{
							RandomizationRangeMin: 20.00,
							RandomizationRangeMax: 50.00,
						},
					},
				},
			},
			{
				Name:        "Transform Full Name",
				Description: "Transforms an existing full name.",
				DataType:    "string",
				Source:      string(TransformFullName),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
						TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{
							PreserveLength: false,
						},
					},
				},
			},
			{
				Name:        "Transform Int64 Phone Number",
				Description: "Transforms an existing phone number that is typed as an integer",
				DataType:    "int64",
				Source:      string(TransformInt64PhoneNumber),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig{
						TransformInt64PhoneNumberConfig: &mgmtv1alpha1.TransformInt64PhoneNumber{
							PreserveLength: false,
						},
					},
				},
			},
			{
				Name:        "Transform Int64",
				Description: "Transforms an existing integer value.",
				DataType:    "int64",
				Source:      string(TransformInt64),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{
						TransformInt64Config: &mgmtv1alpha1.TransformInt64{
							RandomizationRangeMin: 20,
							RandomizationRangeMax: 50,
						},
					},
				},
			},
			{
				Name:        "Transform Last Name",
				Description: "Transforms an existing last name.",
				DataType:    "string",
				Source:      string(TransformLastName),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformLastNameConfig{
						TransformLastNameConfig: &mgmtv1alpha1.TransformLastName{
							PreserveLength: false,
						},
					},
				},
			},
			{
				Name:        "Transform Phone Number",
				Description: "Transforms an existing phone number that is typed as a string.",
				DataType:    "string",
				Source:      string(TransformPhoneNumber),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig{
						TransformPhoneNumberConfig: &mgmtv1alpha1.TransformPhoneNumber{
							PreserveLength: false,
							IncludeHyphens: false,
						},
					},
				},
			},
			{
				Name:        "Transform String",
				Description: "Transforms an existing string value.",
				DataType:    "string",
				Source:      string(TransformString),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
						TransformStringConfig: &mgmtv1alpha1.TransformString{
							PreserveLength: false,
						},
					},
				},
			},
			{
				Name:        "Passthrough",
				Description: "Passes the input value through to the desination with no changes.",
				DataType:    "string",
				Source:      string(Passthrough),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
						PassthroughConfig: &mgmtv1alpha1.Passthrough{},
					},
				},
			},
			{
				Name:        "Null",
				Description: "Inserts a <null> string instead of the source value.",
				DataType:    "string",
				Source:      string(Null),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
						Nullconfig: &mgmtv1alpha1.Null{},
					},
				},
			},
		},
	}), nil
}

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
		return connect.NewResponse(&mgmtv1alpha1.GetUserDefinedTransformerByIdResponse{}), nil
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
	logger = logger.With("transformer", req.Msg.TransformerId)

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
		logger.Info("destination not found")
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

	UserDefinedTransformer := &db_queries.UpdateUserDefinedTransformerParams{
		Name:              req.Msg.Name,
		Description:       req.Msg.Description,
		TransformerConfig: &pg_models.TransformerConfigs{},
		UpdatedByID:       *userUuid,
		ID:                tUuid,
	}

	err = UserDefinedTransformer.TransformerConfig.FromTransformerConfigDto(req.Msg.TransformerConfig)
	if err != nil {
		return nil, err
	}

	t, err := s.db.Q.UpdateUserDefinedTransformer(ctx, s.db.Db, *UserDefinedTransformer)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.UpdateUserDefinedTransformerResponse{
		Transformer: dtomaps.ToUserDefinedTransformerDto(&t),
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
