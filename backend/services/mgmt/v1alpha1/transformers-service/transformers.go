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
				Id:          "80e35b16-d2bc-415a-b63b-e558ad20e5ea",
				Name:        "Generate Email",
				Description: "Generates a new randomized email address.",
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
				Id:          "7c2a7496-4ab3-45af-9917-22731edd14b8",
				Name:        "Generate Realistic Email",
				Description: "Generates a new realistic email address.",
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
				Id:          "063d57f4-674b-44f6-a5e1-acbca0d3031f",
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
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "97793059-6f23-4494-b382-5004f0a3e5ca",
				Name:        "Generate Boolean",
				Description: "Generates a boolean value at random.",
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
				Id:          "3da613e5-cefb-42cb-b5b1-e58e9cfc47d3",
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
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "e7633214-9bc7-4615-8ccd-cf69440294b9",
				Name:        "Generate City",
				Description: "Randomly selects a city from a list of predfined US cities.",
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
				Id:          "d43cf610-58cc-4792-94c7-c6a90c3fc212",
				Name:        "Generate E164 Phone Number",
				Description: "Generates a Generate phone number in e164 format.",
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
				Id:          "3cb1c828-2b1e-4ce4-8402-9e8c70feb84b",
				Name:        "Generate First Name",
				Description: "Generates a random first name. ",
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
				Id:          "ff371ab8-89f8-4308-bdd9-099842a68bb5",
				Name:        "Generate Float64",
				Description: "Generates a random float64 value.",
				DataType:    "float64",
				Source:      string(GenerateFloat),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateFloatConfig{
						GenerateFloatConfig: &mgmtv1alpha1.GenerateFloat{
							Sign:                "positive",
							DigitsBeforeDecimal: 3,
							DigitsAfterDecimal:  3,
						},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "e2b73154-f90f-4ef4-bf36-870b123e4192",
				Name:        "Generate Full Address",
				Description: "Randomly generates a street address in the format: {street_num} {street_addresss} {street_descriptor} {city}, {state} {zipcode}. For example, 123 Main Street Boston, Massachusetts 02169.",
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
				Id:          "f9c36ccb-53ff-4c77-960b-c2e18bec1617",
				Name:        "Generate Full Name",
				Description: "Generates a new full name consisting of a first and last name",
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
				Id:          "c0028c26-25b9-4366-953f-2b6e20e420cc",
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
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "20a63adc-34ba-4845-b354-587791749595",
				Name:        "Generate int64 Phone Number",
				Description: "Generates a new phone number of type int64 with a default length of 10.",
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
				Id:          "0b591b7c-859d-4d55-9a52-9d3526ee1345",
				Name:        "Generate Random Int64",
				Description: "Generates a random integer value with a default length of 4 unless the Integer Length or Preserve Length paramters are defined.", DataType: "int64",
				Source: string(GenerateInt),
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
				Id:          "2e28df1d-3809-4b8b-aaf9-80723cc1c939",
				Name:        "Generate Last Name",
				Description: "Generates a random last name.", DataType: "int64",
				Source: string(GenerateLastName),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig{
						GenerateLastNameConfig: &mgmtv1alpha1.GenerateLastName{},
					},
				},
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "3f54ef47-8415-4ea5-ab9a-bd2777c64a72",
				Name:        "Generate SHA256 Hash",
				Description: "SHA256 hashes a randomly generated value.",
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
				Id:          "5e6bb23b-5ea7-4feb-a204-5fe7c84e57e3",
				Name:        "Generate SSN",
				Description: "Generates a completely random social security numbers including the hyphens in the format <xxx-xx-xxxx>",
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
				Id:          "b67e04c8-a550-4864-a6e4-c17e79882df1",
				Name:        "Generate State",
				Description: "Randomly selects a US state and returns the two-character state code.",
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
				Id:          "e0d1b3eb-ac98-42a7-935d-f2d02ad1767b",
				Name:        "Generate Street Address",
				Description: "Randomly generates a street address in the format: {street_num} {street_addresss} {street_descriptor}. For example, 123 Main Street.",
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
				Id:          "bcc99015-49d9-46eb-9040-32bf9e0a711a",
				Name:        "Generate String Phone Number",
				Description: "Generates a Generate phone number and returns it as a string.",
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
				Name:        "Generate Random String",
				Description: "Creates a randomly ordered alphanumeric string with a default length of 10 unless the String Length parameter are defined.",
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
				Id:          "d4df826a-767e-4f32-beeb-fed309162ac6",
				Name:        "Generate Unix Timestamp",
				Description: "Randomly generates a Unix timestamp",
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
				Id:          "bc3b1394-239d-4e98-ba6a-60f68b2023a6",
				Name:        "Generate Username",
				Description: "Randomly generates a username in the format<first_initial><last_name>.",
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
				Id:          "012e5683-1669-4190-b0d3-0173ce1561d4",
				Name:        "Generate UTC Timestamp",
				Description: "Randomly generates a UTC timestamp.",
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
				Id:          "a9d384ee-1095-4a8d-9668-bcc3aaf4291f",
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
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "ddbf48b2-02da-409a-83c4-6e5cd6317079",
				Name:        "Generate Zipcode",
				Description: "Randomly selects a zip code from a list of predefined US zipcodes.",
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
				Id:          "5d8a6030-fed5-42af-b70f-fc40825e0df6",
				Name:        "Transform E164 Phone Number",
				Description: "Transforms an existing E164 formatted phone number.",
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
				Id:          "cf4230f8-d8fd-450f-8829-10b8643362ac",
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
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "5dc13fc2-349c-497a-ba52-e0880a9bce1d",
				Name:        "Transform Float64",
				Description: "Transforms an existing float value.",
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
				Id:          "5c9ef9e2-4753-4603-9d68-397e92dd943f",
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
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "b8688369-f3b3-40d7-9b47-d9e8fcbed91e",
				Name:        "Transform Int64 Phone Number",
				Description: "Transforms an existing phone number that is typed as an integer",
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
				Id:          "7375aece-16ba-4f32-8629-e88f3430631d",
				Name:        "Transform Int64",
				Description: "Transforms an existing integer value.",
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
				Id:          "5b53b065-eb1c-451c-aa74-d04a3dafc286",
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
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "0265c84e-bfd5-4a15-a747-890faaa85e7e",
				Name:        "Transform Phone Number",
				Description: "Transforms an existing phone number that is typed as a string.",
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
				Id:          "de7db429-5cf1-465a-b76b-ec2c9c77cafe",
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
				CreatedAt: timestampNow,
				UpdatedAt: timestampNow,
			},
			{
				Id:          "d5a1492a-9d4e-40c8-b482-6ab04c7ce096",
				Name:        "Passthrough",
				Description: "Passes the input value through to the desination with no changes.",
				DataType:    "string",
				Source:      string(Passthrough),
				Config:      &mgmtv1alpha1.TransformerConfig{},
				CreatedAt:   timestampNow,
				UpdatedAt:   timestampNow,
			},
			{
				Id:          "4b77908f-591c-40df-bf4f-e124eb1a00a2",
				Name:        "Null",
				Description: "Inserts a <null> string instead of the source value.",
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
