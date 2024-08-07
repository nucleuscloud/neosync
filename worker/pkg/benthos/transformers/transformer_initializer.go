package transformers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
)

type TransformerExecutor struct {
	Opts   any
	Mutate func(value any, opts any) (any, error)
}

func InitializeTransformer(transformerMapping *mgmtv1alpha1.JobMappingTransformer) (*TransformerExecutor, error) {
	maxLength := int64(10000) // TODO: update this based on colInfo if available
	switch transformerMapping.Source {
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CATEGORICAL:
		categories := transformerMapping.Config.GetGenerateCategoricalConfig().Categories
		opts, err := NewGenerateCategoricalOpts(categories)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateCategorical().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL:
		opts, err := NewGenerateBoolOpts(nil)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateBool().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_STRING:
		pl := transformerMapping.Config.GetTransformStringConfig().PreserveLength
		minLength := int64(3) // TODO: pull this value from the database schema
		opts, err := NewTransformStringOpts(&pl, &minLength, &maxLength)
		if err != nil {
			return nil, err
		}
		transform := NewTransformString().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64:
		rMin := transformerMapping.Config.GetTransformInt64Config().RandomizationRangeMin
		rMax := transformerMapping.Config.GetTransformInt64Config().RandomizationRangeMax
		opts, err := NewTransformInt64Opts(rMin, rMax)
		if err != nil {
			return nil, err
		}
		transform := NewTransformInt64().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FULL_NAME:
		pl := transformerMapping.Config.GetTransformFullNameConfig().PreserveLength
		opts, err := NewTransformFullNameOpts(nil, &pl, &maxLength)
		if err != nil {
			return nil, err
		}
		transform := NewTransformFullName().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_EMAIL:
		emailType := transformerMapping.GetConfig().GetGenerateEmailConfig().GetEmailType()
		if emailType == mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_UNSPECIFIED {
			emailType = mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_UUID_V4
		}
		emailStrType := emailType.String()
		opts, err := NewGenerateEmailOpts(&maxLength, &emailStrType, nil)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateEmail().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_EMAIL:
		config := transformerMapping.Config.GetTransformEmailConfig()
		emailTypeStr := config.EmailType.String()
		invalidEmailActionStr := config.InvalidEmailAction.String()
		var excludedDomains any = config.ExcludedDomains
		opts, err := NewTransformEmailOpts(
			&config.PreserveDomain,
			&config.PreserveLength,
			&excludedDomains,
			&maxLength,
			nil,
			&emailTypeStr,
			&invalidEmailActionStr,
		)
		if err != nil {
			return nil, err
		}
		transform := NewTransformEmail().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CARD_NUMBER:
		luhn := transformerMapping.Config.GetGenerateCardNumberConfig().ValidLuhn
		opts, err := NewGenerateCardNumberOpts(luhn)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateCardNumber().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CITY:
		opts, err := NewGenerateCityOpts(maxLength)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateCity().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_E164_PHONE_NUMBER:
		minValue := transformerMapping.Config.GetGenerateE164PhoneNumberConfig().Min
		maxValue := transformerMapping.Config.GetGenerateE164PhoneNumberConfig().Max
		opts, err := NewGenerateInternationalPhoneNumberOpts(minValue, maxValue)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateInternationalPhoneNumber().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FIRST_NAME:
		opts, err := NewGenerateFirstNameOpts(&maxLength, nil)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateFirstName().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FLOAT64:
		config := transformerMapping.Config.GetGenerateFloat64Config()
		opts, err := NewGenerateFloat64Opts(
			&config.RandomizeSign,
			config.Min,
			config.Max,
			&config.Precision,
			nil, // TODO: update scale based on colInfo if available
			nil,
		)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateFloat64().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_ADDRESS:
		opts, err := NewGenerateFullAddressOpts(maxLength)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateFullAddress().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME:
		opts, err := NewGenerateFullNameOpts(&maxLength, nil)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateFullName().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_GENDER:
		ab := transformerMapping.Config.GetGenerateGenderConfig().Abbreviate
		opts, err := NewGenerateGenderOpts(&ab, &maxLength, nil)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateGender().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64_PHONE_NUMBER:
		opts, err := NewGenerateInt64PhoneNumberOpts()
		if err != nil {
			return nil, err
		}
		generate := NewGenerateInt64PhoneNumber().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64:
		config := transformerMapping.Config.GetGenerateInt64Config()
		opts, err := NewGenerateInt64Opts(&config.RandomizeSign, config.Min, config.Max)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateInt64().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_LAST_NAME:
		opts, err := NewGenerateLastNameOpts(&maxLength, nil)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateLastName().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SHA256HASH:
		opts, err := NewGenerateSHA256HashOpts()
		if err != nil {
			return nil, err
		}
		generate := NewGenerateSHA256Hash().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SSN:
		opts, err := NewGenerateSSNOpts(nil)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateSSN().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STATE:
		generateFullName := transformerMapping.Config.GetGenerateStateConfig().GenerateFullName
		opts, err := NewGenerateStateOpts(&generateFullName)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateState().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STREET_ADDRESS:
		opts, err := NewGenerateStreetAddressOpts(maxLength)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateStreetAddress().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STRING_PHONE_NUMBER:
		minValue := transformerMapping.Config.GetGenerateStringPhoneNumberConfig().Min
		maxValue := transformerMapping.Config.GetGenerateStringPhoneNumberConfig().Max
		minValue = transformer_utils.MinInt(minValue, maxLength)
		maxValue = transformer_utils.Ceil(maxValue, maxLength)
		opts, err := NewGenerateStringPhoneNumberOpts(minValue, maxValue)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateStringPhoneNumber().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_RANDOM_STRING:
		config := transformerMapping.Config.GetGenerateStringConfig()
		opts, err := NewGenerateRandomStringOpts(config.Min, config.Max)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateRandomString().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UNIXTIMESTAMP:
		opts, err := NewGenerateUnixTimestampOpts()
		if err != nil {
			return nil, err
		}
		generate := NewGenerateUnixTimestamp().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_USERNAME:
		opts, err := NewGenerateUsernameOpts(&maxLength, nil)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateUsername().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UTCTIMESTAMP:
		opts, err := NewGenerateUTCTimestampOpts()
		if err != nil {
			return nil, err
		}
		generate := NewGenerateUTCTimestamp().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UUID:
		ih := transformerMapping.Config.GetGenerateUuidConfig().IncludeHyphens
		opts, err := NewGenerateUUIDOpts(&ih)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateUUID().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_ZIPCODE:
		opts, err := NewGenerateZipcodeOpts()
		if err != nil {
			return nil, err
		}
		generate := NewGenerateZipcode().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_E164_PHONE_NUMBER:
		config := transformerMapping.Config.GetTransformE164PhoneNumberConfig()
		opts, err := NewTransformE164PhoneNumberOpts(config.PreserveLength, &maxLength)
		if err != nil {
			return nil, err
		}
		transform := NewTransformE164PhoneNumber().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FIRST_NAME:
		config := transformerMapping.Config.GetTransformFirstNameConfig()
		opts, err := NewTransformFirstNameOpts(&maxLength, &config.PreserveLength, nil)
		if err != nil {
			return nil, err
		}
		transform := NewTransformFirstName().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FLOAT64:
		config := transformerMapping.Config.GetTransformFloat64Config()
		opts, err := NewTransformFloat64Opts(
			config.RandomizationRangeMin,
			config.RandomizationRangeMax,
			nil, // TODO: update precision based on colInfo if available
			nil, // TODO: update scale based on colInfo if available
			nil,
		)
		if err != nil {
			return nil, err
		}
		transform := NewTransformFloat64().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64_PHONE_NUMBER:
		config := transformerMapping.Config.GetTransformInt64PhoneNumberConfig()
		opts, err := NewTransformInt64PhoneNumberOpts(config.PreserveLength)
		if err != nil {
			return nil, err
		}
		transform := NewTransformInt64PhoneNumber().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_LAST_NAME:
		config := transformerMapping.Config.GetTransformLastNameConfig()
		opts, err := NewTransformLastNameOpts(&maxLength, &config.PreserveLength, nil)
		if err != nil {
			return nil, err
		}
		transform := NewTransformLastName().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_PHONE_NUMBER:
		config := transformerMapping.Config.GetTransformPhoneNumberConfig()
		opts, err := NewTransformStringPhoneNumberOpts(config.PreserveLength, maxLength)
		if err != nil {
			return nil, err
		}
		transform := NewTransformStringPhoneNumber().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL:
		return &TransformerExecutor{
			Opts: nil,
			Mutate: func(value any, opts any) (any, error) {
				return "null", nil
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT:
		return &TransformerExecutor{
			Opts: nil,
			Mutate: func(value any, opts any) (any, error) {
				return `"DEFAULT"`, nil
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_CHARACTER_SCRAMBLE:
		config := transformerMapping.Config.GetTransformCharacterScrambleConfig()
		opts, err := NewTransformCharacterScrambleOpts(config.UserProvidedRegex)
		if err != nil {
			return nil, err
		}
		transform := NewTransformCharacterScramble().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported transformer: %v", transformerMapping.Source)
	}
}
