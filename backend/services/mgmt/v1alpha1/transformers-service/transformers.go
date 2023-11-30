package v1alpha1_transformersservice

import (
	"context"
	"time"

	"connectrpc.com/connect"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	logger_interceptor "github.com/nucleuscloud/neosync/backend/internal/connect/interceptors/logger"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	timestampNow := timestamppb.New(time.Now())

	return connect.NewResponse(&mgmtv1alpha1.GetSystemTransformersResponse{
		Transformers: []*mgmtv1alpha1.Transformer{
			{
				Id:          "",
				Name:        string(GenerateEmail),
				Description: "Generates a new Generate email address.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateEmail),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
						GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateRealisticEmail),
				Description: "Generates a new realistic email address.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateRealisticEmail),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateRealisticEmailConfig{
						GenerateRealisticEmailConfig: &mgmtv1alpha1.GenerateRealisticEmail{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(TransformEmail),
				Description: "Transforms an existing email address..",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
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
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateBool),
				Description: "Generates a boolean value at random.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "boolean",
				Source:      string(GenerateBool),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
						GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateCardNumber),
				Description: "Generates a card number.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "int64",
				Source:      string(GenerateCardNumber),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{
						GenerateCardNumberConfig: &mgmtv1alpha1.GenerateCardNumber{
							ValidLuhn: true,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateCity),
				Description: "Randomly selects a city from a list of predfined US cities.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateCity),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
						GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateE164Number),
				Description: "Generates a Generate phone number in e164 format.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateE164Number),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateE164NumberConfig{
						GenerateE164NumberConfig: &mgmtv1alpha1.GenerateE164Number{
							Length: 12,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateFirstName),
				Description: "Generates a random first name. ",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateFirstName),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
						GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateFloat),
				Description: "Generates a random float64 value.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "float64",
				Source:      string(GenerateFloat),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateFloatConfig{
						GenerateFloatConfig: &mgmtv1alpha1.GenerateFloat{
							Sign:                "postiive",
							DigitsBeforeDecimal: 3,
							DigitsAfterDecimal:  3,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateFullAddress),
				Description: "Randomly generates a street address in the format: {street_num} {street_addresss} {street_descriptor} {city}, {state} {zipcode}. For example, 123 Main Street Boston, Massachusetts 02169.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateFullAddress),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig{
						GenerateFullAddressConfig: &mgmtv1alpha1.GenerateFullAddress{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateFullName),
				Description: "Generates a new full name consisting of a first and last name",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateFullName),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
						GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateGender),
				Description: "Randomly generates one of the following genders: female, male, undefined, nonbinary.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateGender),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateGenderConfig{
						GenerateGenderConfig: &mgmtv1alpha1.GenerateGender{
							Abbreviate: false,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateInt64Phone),
				Description: "Generates a new phone number of type int64 with a default length of 10.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "int64",
				Source:      string(GenerateInt64Phone),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneConfig{
						GenerateInt64PhoneConfig: &mgmtv1alpha1.GenerateInt64Phone{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateInt),
				Description: "Generates a random integer value with a default length of 4 unless the Integer Length or Preserve Length paramters are defined.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "int64",
				Source:      string(GenerateInt),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateIntConfig{
						GenerateIntConfig: &mgmtv1alpha1.GenerateInt{
							Length: 4,
							Sign:   "positive",
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateLastName),
				Description: "Generates a random last name.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "int64",
				Source:      string(GenerateLastName),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig{
						GenerateLastNameConfig: &mgmtv1alpha1.GenerateLastName{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateShaHash256),
				Description: "SHA256 hashes a randomly generated value.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateShaHash256),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig{
						GenerateSha256HashConfig: &mgmtv1alpha1.GenerateSha256Hash{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateSSN),
				Description: "Generates a completely random social security numbers including the hyphens in the format <xxx-xx-xxxx>",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateSSN),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
						GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateState),
				Description: "Randomly selects a US state and returns the two-character state code.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateState),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateStateConfig{
						GenerateStateConfig: &mgmtv1alpha1.GenerateState{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateStreetAddress),
				Description: "Randomly generates a street address in the format: {street_num} {street_addresss} {street_descriptor}. For example, 123 Main Street.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateStreetAddress),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig{
						GenerateStreetAddressConfig: &mgmtv1alpha1.GenerateStreetAddress{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateStringPhone),
				Description: "Generates a Generate phone number and returns it as a string.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateStringPhone),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateStringPhoneConfig{
						GenerateStringPhoneConfig: &mgmtv1alpha1.GenerateStringPhone{
							E164Format:     false,
							IncludeHyphens: false,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateString),
				Description: "Creates a randomly ordered alphanumeric string with a default length of 10 unless the String Length parameter are defined.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateString),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
						GenerateStringConfig: &mgmtv1alpha1.GenerateString{
							Length: 6,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateUnixTimestamp),
				Description: "Randomly generates a Unix timestamp",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "int64",
				Source:      string(GenerateUnixTimestamp),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{
						GenerateUnixtimestampConfig: &mgmtv1alpha1.GenerateUnixTimestamp{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateUsername),
				Description: "Randomly generates a username in the format<first_initial><last_name>.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateUsername),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig{
						GenerateUsernameConfig: &mgmtv1alpha1.GenerateUsername{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateUtcTimestamp),
				Description: "Randomly generates a UTC timestamp.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "time",
				Source:      string(GenerateUtcTimestamp),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig{
						GenerateUtctimestampConfig: &mgmtv1alpha1.GenerateUtcTimestamp{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateUuid),
				Description: "Generates a new UUIDv4 id.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "uuid",
				Source:      string(GenerateUuid),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
						GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
							IncludeHyphens: true,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(GenerateZipcode),
				Description: "Randomly selects a zip code from a list of predefined US zipcodes.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(GenerateZipcode),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig{
						GenerateZipcodeConfig: &mgmtv1alpha1.GenerateZipcode{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow},
			{
				Id:          "",
				Name:        string(TransformE164Phone),
				Description: "Transforms an existing E164 formatted phone number.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(TransformE164Phone),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformE164PhoneConfig{
						TransformE164PhoneConfig: &mgmtv1alpha1.TransformE164Phone{
							PreserveLength: false,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(TransformFirstName),
				Description: "Transforms an existing first name",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(TransformFirstName),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
						TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{
							PreserveLength: false,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(TransformFloat),
				Description: "Transforms an existing float value.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "float64",
				Source:      string(TransformFloat),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformFloatConfig{
						TransformFloatConfig: &mgmtv1alpha1.TransformFloat{
							PreserveLength: false,
							PreserveSign:   true,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(TransformFullName),
				Description: "Transforms an existing full name.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(TransformFullName),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
						TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{
							PreserveLength: false,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(TransformIntPhone),
				Description: "Transforms an existing phone number that is typed as an integer",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "int64",
				Source:      string(TransformIntPhone),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformIntPhoneConfig{
						TransformIntPhoneConfig: &mgmtv1alpha1.TransformIntPhone{
							PreserveLength: false,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(TransformInt),
				Description: "Transforms an existing integer value.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "int64",
				Source:      string(TransformInt),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformIntConfig{
						TransformIntConfig: &mgmtv1alpha1.TransformInt{
							PreserveLength: false,
							PreserveSign:   true,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(TransformLastName),
				Description: "Transforms an existing last name.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(TransformLastName),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformLastNameConfig{
						TransformLastNameConfig: &mgmtv1alpha1.TransformLastName{
							PreserveLength: false,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(TransformPhone),
				Description: "Transforms an existing phone number that is typed as a string.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(TransformPhone),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformPhoneConfig{
						TransformPhoneConfig: &mgmtv1alpha1.TransformPhone{
							PreserveLength: false,
							IncludeHyphens: false,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(TransformString),
				Description: "Transforms an existing string value.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(TransformString),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
						TransformStringConfig: &mgmtv1alpha1.TransformString{
							PreserveLength: false,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "",
				Name:        string(Passthrough),
				Description: "Passes the input value through to the desination with no changes.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(Passthrough),
				Config:      &mgmtv1alpha1.TransformerConfig{},
				CreatedAt:   timestampNow,
				UpdatedAt:   timestampNow,
			},
			{
				Id:          "",
				Name:        string(Null),
				Description: "Inserts a <null> string instead of the source value.",
				Type:        mgmtv1alpha1.TransformerType_TRANSFORMER_TYPE_SYSTEM,
				DataType:    "string",
				Source:      string(Null),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
						Nullconfig: &mgmtv1alpha1.Null{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
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

	dtoTransformers := []*mgmtv1alpha1.Transformer{}
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
