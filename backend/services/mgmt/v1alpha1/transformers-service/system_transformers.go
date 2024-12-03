package v1alpha1_transformersservice

import (
	"cmp"
	"context"
	"slices"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	ee_transformers "github.com/nucleuscloud/neosync/internal/ee/transformers"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
)

var (
	defaultGenerateEmailType = mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_UUID_V4

	defaultInvalidEmailAction = mgmtv1alpha1.InvalidEmailAction_INVALID_EMAIL_ACTION_REJECT

	defaultGenerateIpv4 = mgmtv1alpha1.GenerateIpAddressType_GENERATE_IP_ADDRESS_TYPE_V4_PUBLIC

	// base transformers
	baseSystemTransformers = []*mgmtv1alpha1.SystemTransformer{
		{

			Name:              "Generate Email",
			Description:       "Generates a new randomized email address.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_EMAIL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
					GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{
						EmailType: &defaultGenerateEmailType,
					},
				},
			},
		},
		{
			Name:              "Transform Email",
			Description:       "Transforms an existing email address.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_EMAIL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
					TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
						PreserveDomain:     gotypeutil.ToPtr(false),
						PreserveLength:     gotypeutil.ToPtr(false),
						ExcludedDomains:    []string{},
						EmailType:          &defaultGenerateEmailType,
						InvalidEmailAction: &defaultInvalidEmailAction,
					},
				},
			},
		},
		{
			Name:              "Generate Boolean",
			Description:       "Generates a boolean value at random.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_BOOLEAN,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_BOOLEAN, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
					GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
				},
			},
		},
		{
			Name:              "Generate Card Number",
			Description:       "Generates a card number.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_INT64,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_INT64, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CARD_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{
					GenerateCardNumberConfig: &mgmtv1alpha1.GenerateCardNumber{
						ValidLuhn: gotypeutil.ToPtr(true),
					},
				},
			},
		},
		{
			Name:              "Generate City",
			Description:       "Randomly selects a city from a list of predfined US cities.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CITY,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
					GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
				},
			},
		},
		{
			Name:              "Use Column Default",
			Description:       "Defers to the database column default",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_ANY, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{
					GenerateDefaultConfig: &mgmtv1alpha1.GenerateDefault{},
				},
			},
		},
		{
			Name:              "Generate International Phone Number",
			Description:       "Generates a phone number in international format with the + character at the start of the phone number. Note that the + sign is not included in the min or max.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_E164_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateE164PhoneNumberConfig{
					GenerateE164PhoneNumberConfig: &mgmtv1alpha1.GenerateE164PhoneNumber{
						Min: gotypeutil.ToPtr(int64(9)),
						Max: gotypeutil.ToPtr(int64(15)),
					},
				},
			},
		},
		{
			Name:              "Generate First Name",
			Description:       "Generates a random first name. ",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FIRST_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
					GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
				},
			},
		},
		{
			Name:              "Generate Float64",
			Description:       "Generates a random float64 value.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_FLOAT64,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_FLOAT64, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FLOAT64,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{
					GenerateFloat64Config: &mgmtv1alpha1.GenerateFloat64{
						RandomizeSign: gotypeutil.ToPtr(false),
						Min:           gotypeutil.ToPtr(1.00),
						Max:           gotypeutil.ToPtr(100.00),
						Precision:     gotypeutil.ToPtr(int64(6)),
					},
				},
			},
		},
		{
			Name:              "Generate Full Address",
			Description:       "Randomly generates a street address in the format: {street_num} {street_addresss} {street_descriptor} {city}, {state} {zipcode}. For example, 123 Main Street Boston, Massachusetts 02169.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_ADDRESS,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig{
					GenerateFullAddressConfig: &mgmtv1alpha1.GenerateFullAddress{},
				},
			},
		},
		{
			Name:              "Generate Full Name",
			Description:       "Generates a new full name consisting of a first and last name",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		},
		{
			Name:              "Generate Gender",
			Description:       "Randomly generates one of the following genders: female, male, undefined, nonbinary.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_GENDER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateGenderConfig{
					GenerateGenderConfig: &mgmtv1alpha1.GenerateGender{
						Abbreviate: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Name:              "Generate Int64 Phone Number",
			Description:       "Generates a new phone number with a default length of 10.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_INT64,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_INT64, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneNumberConfig{
					GenerateInt64PhoneNumberConfig: &mgmtv1alpha1.GenerateInt64PhoneNumber{},
				},
			},
		},
		{
			Name:              "Generate Random Int64",
			Description:       "Generates a random int64 value.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_INT64,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_INT64, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
					GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
						RandomizeSign: gotypeutil.ToPtr(false),
						Min:           gotypeutil.ToPtr(int64(1)),
						Max:           gotypeutil.ToPtr(int64(40)),
					},
				},
			},
		},
		{
			Name:              "Generate Last Name",
			Description:       "Generates a random last name.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_LAST_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig{
					GenerateLastNameConfig: &mgmtv1alpha1.GenerateLastName{},
				},
			},
		},
		{
			Name:              "Generate SHA256 Hash",
			Description:       "SHA256 hashes a randomly generated value.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SHA256HASH,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig{
					GenerateSha256HashConfig: &mgmtv1alpha1.GenerateSha256Hash{},
				},
			},
		},
		{
			Name:              "Generate SSN",
			Description:       "Generates a completely random social security numbers including the hyphens in the format <xxx-xx-xxxx>",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SSN,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
					GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
				},
			},
		},
		{
			Name:              "Generate State",
			Description:       "Randomly selects a US state.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STATE,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStateConfig{
					GenerateStateConfig: &mgmtv1alpha1.GenerateState{
						GenerateFullName: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Name:              "Generate Street Address",
			Description:       "Randomly generates a street address in the format: {street_num} {street_addresss} {street_descriptor}. For example, 123 Main Street.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STREET_ADDRESS,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig{
					GenerateStreetAddressConfig: &mgmtv1alpha1.GenerateStreetAddress{},
				},
			},
		},
		{
			Name:              "Generate String Phone Number",
			Description:       "Generates a phone number and returns it as a string.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STRING_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig{
					GenerateStringPhoneNumberConfig: &mgmtv1alpha1.GenerateStringPhoneNumber{
						Min: gotypeutil.ToPtr(int64(9)),
						Max: gotypeutil.ToPtr(int64(14)),
					},
				},
			},
		},
		{
			Name:              "Generate Random String",
			Description:       "Creates a randomly ordered alphanumeric string between the specified range",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_RANDOM_STRING,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
					GenerateStringConfig: &mgmtv1alpha1.GenerateString{
						Min: gotypeutil.ToPtr(int64(2)),
						Max: gotypeutil.ToPtr(int64(7)),
					},
				},
			},
		},
		{
			Name:              "Generate Unix Timestamp",
			Description:       "Randomly generates a Unix timestamp",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_INT64,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_INT64, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UNIXTIMESTAMP,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{
					GenerateUnixtimestampConfig: &mgmtv1alpha1.GenerateUnixTimestamp{},
				},
			},
		},
		{
			Name:              "Generate Username",
			Description:       "Randomly generates a username in the format <first_initial><last_name>.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_USERNAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig{
					GenerateUsernameConfig: &mgmtv1alpha1.GenerateUsername{},
				},
			},
		},
		{
			Name:              "Generate UTC Timestamp",
			Description:       "Randomly generates a UTC timestamp.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_TIME,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_TIME, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UTCTIMESTAMP,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig{
					GenerateUtctimestampConfig: &mgmtv1alpha1.GenerateUtcTimestamp{},
				},
			},
		},
		{
			Name:              "Generate UUID",
			Description:       "Generates a new UUIDv4 identifier.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_UUID,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_UUID, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UUID,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
					GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
						IncludeHyphens: gotypeutil.ToPtr(true),
					},
				},
			},
		},
		{
			Name:              "Generate Zipcode",
			Description:       "Randomly selects a zip code from a list of predefined US zipcodes.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_ZIPCODE,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig{
					GenerateZipcodeConfig: &mgmtv1alpha1.GenerateZipcode{},
				},
			},
		},
		{
			Name:              "Transform E164 Phone Number",
			Description:       "Transforms an existing E164 formatted phone number.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_E164_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig{
					TransformE164PhoneNumberConfig: &mgmtv1alpha1.TransformE164PhoneNumber{
						PreserveLength: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Name:              "Transform First Name",
			Description:       "Transforms an existing first name",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FIRST_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
					TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{
						PreserveLength: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Name:              "Transform Float64",
			Description:       "Transforms an existing float value.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_FLOAT64,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_FLOAT64, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FLOAT64,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFloat64Config{
					TransformFloat64Config: &mgmtv1alpha1.TransformFloat64{
						RandomizationRangeMin: gotypeutil.ToPtr(20.00),
						RandomizationRangeMax: gotypeutil.ToPtr(50.00),
					},
				},
			},
		},
		{
			Name:              "Transform Full Name",
			Description:       "Transforms an existing full name.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FULL_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
					TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{
						PreserveLength: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Name:              "Transform Int64 Phone Number",
			Description:       "Transforms an existing phone number that is typed as an integer",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_INT64,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_INT64, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig{
					TransformInt64PhoneNumberConfig: &mgmtv1alpha1.TransformInt64PhoneNumber{
						PreserveLength: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Name:              "Transform Int64",
			Description:       "Transforms an existing integer value.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_INT64,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_INT64, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{
					TransformInt64Config: &mgmtv1alpha1.TransformInt64{
						RandomizationRangeMin: gotypeutil.ToPtr(int64(20)),
						RandomizationRangeMax: gotypeutil.ToPtr(int64(50)),
					},
				},
			},
		},
		{
			Name:              "Transform Last Name",
			Description:       "Transforms an existing last name.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_LAST_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformLastNameConfig{
					TransformLastNameConfig: &mgmtv1alpha1.TransformLastName{
						PreserveLength: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Name:              "Transform String Phone Number",
			Description:       "Transforms an existing phone number that is typed as a string.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig{
					TransformPhoneNumberConfig: &mgmtv1alpha1.TransformPhoneNumber{
						PreserveLength: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Name:              "Transform String",
			Description:       "Transforms an existing string value.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_STRING,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
					TransformStringConfig: &mgmtv1alpha1.TransformString{
						PreserveLength: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Name:              "Passthrough",
			Description:       "Passes the input value through to the desination with no changes.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_ANY,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_ANY, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
					PassthroughConfig: &mgmtv1alpha1.Passthrough{},
				},
			},
		},
		{
			Name:              "Null",
			Description:       "Inserts a <null> string instead of the source value.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
					Nullconfig: &mgmtv1alpha1.Null{},
				},
			},
		},
		{
			Name:              "Transform Javascript",
			Description:       "Write custom javascript to transform data",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_ANY,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_ANY, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
					TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{Code: `return value + "test";`},
				},
			},
		},
		{
			Name:              "Generate Categorical",
			Description:       "Randomly selects a value from a predefined list of values",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CATEGORICAL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig{
					GenerateCategoricalConfig: &mgmtv1alpha1.GenerateCategorical{
						Categories: gotypeutil.ToPtr("value1,value2"),
					},
				},
			},
		},
		{
			Name:              "Transform Character Scramble",
			Description:       "Transforms a string value by scrambling each character with another character in the same unicode block. Letters will be substituted with letters, numbers with numbers and special characters with special characters. Spaces and capitalization is preserved.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_CHARACTER_SCRAMBLE,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
					TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{
						UserProvidedRegex: nil,
					},
				},
			},
		},
		{
			Name:              "Generate Javascript",
			Description:       "Write custom Javascript to generate synthetic data.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_ANY,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_ANY, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
					GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{Code: `return "testvalue";`},
				},
			},
		},
		{
			Name:              "Generate Country",
			Description:       "Randomly selects a Country.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_COUNTRY,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCountryConfig{
					GenerateCountryConfig: &mgmtv1alpha1.GenerateCountry{
						GenerateFullName: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Name:              "Generate Business Name",
			Description:       "Generates a random business name.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BUSINESS_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateBusinessNameConfig{
					GenerateBusinessNameConfig: &mgmtv1alpha1.GenerateBusinessName{},
				},
			},
		},
		{
			Name:              "Generate IP Address",
			Description:       "Generates a random IP address.",
			DataType:          mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			DataTypes:         []mgmtv1alpha1.TransformerDataType{mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING, mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL},
			SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_GENERATE, mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
			Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_IP_ADDRESS,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateIpAddressConfig{
					GenerateIpAddressConfig: &mgmtv1alpha1.GenerateIpAddress{
						IpType: &defaultGenerateIpv4,
					},
				},
			},
		},
	}

	// base transformers + ee transformers
	allSystemTransformers = []*mgmtv1alpha1.SystemTransformer{}

	baseSystemTransformerSourceMap = map[mgmtv1alpha1.TransformerSource]*mgmtv1alpha1.SystemTransformer{}

	allSystemTransformersSourceMap = map[mgmtv1alpha1.TransformerSource]*mgmtv1alpha1.SystemTransformer{}
)

func init() {
	allSystemTransformers = append(allSystemTransformers, baseSystemTransformers...)
	allSystemTransformers = append(allSystemTransformers, ee_transformers.Transformers...)

	slices.SortFunc(baseSystemTransformers, func(t1, t2 *mgmtv1alpha1.SystemTransformer) int {
		return cmp.Compare(t1.Name, t2.Name)
	})
	slices.SortFunc(allSystemTransformers, func(t1, t2 *mgmtv1alpha1.SystemTransformer) int {
		return cmp.Compare(t1.Name, t2.Name)
	})

	// hydrate the system transformer map when the system boots up
	for _, transformer := range baseSystemTransformers {
		baseSystemTransformerSourceMap[transformer.Source] = transformer
	}
	for _, transformer := range allSystemTransformers {
		allSystemTransformersSourceMap[transformer.Source] = transformer
	}
}

func (s *Service) GetSystemTransformers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetSystemTransformersRequest],
) (*connect.Response[mgmtv1alpha1.GetSystemTransformersResponse], error) {
	systemTransformers := s.getSystemTransformers()
	return connect.NewResponse(&mgmtv1alpha1.GetSystemTransformersResponse{
		Transformers: systemTransformers,
	}), nil
}

func (s *Service) GetSystemTransformerBySource(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetSystemTransformerBySourceRequest],
) (*connect.Response[mgmtv1alpha1.GetSystemTransformerBySourceResponse], error) {
	transformerMap := s.getSystemTransformerSourceMap()

	transformer, ok := transformerMap[req.Msg.GetSource()]
	if !ok {
		return nil, nucleuserrors.NewNotFound("unable to find system transformer with provided source")
	}
	return connect.NewResponse(&mgmtv1alpha1.GetSystemTransformerBySourceResponse{
		Transformer: transformer,
	}), nil
}

func (s *Service) getSystemTransformerSourceMap() map[mgmtv1alpha1.TransformerSource]*mgmtv1alpha1.SystemTransformer {
	if s.cfg.IsNeosyncCloud {
		return allSystemTransformersSourceMap
	}
	return baseSystemTransformerSourceMap
}

func (s *Service) getSystemTransformers() []*mgmtv1alpha1.SystemTransformer {
	if s.cfg.IsNeosyncCloud {
		return allSystemTransformers
	}
	return baseSystemTransformers
}
