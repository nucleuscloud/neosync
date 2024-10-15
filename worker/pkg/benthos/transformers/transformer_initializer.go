package transformers

import (
	"context"
	"errors"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	ee_transformer_fns "github.com/nucleuscloud/neosync/internal/ee/transformers/functions"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
)

type TransformerExecutor struct {
	Opts   any
	Mutate func(value any, opts any) (any, error)
}

type TransformerExecutorOption func(c *TransformerExecutorConfig)

type TransformerExecutorConfig struct {
	transformPiiText *transformPiiTextConfig
}

type transformPiiTextConfig struct {
	analyze   presidioapi.AnalyzeInterface
	anonymize presidioapi.AnonymizeInterface
}

func WithTransformPiiTextConfig(analyze presidioapi.AnalyzeInterface, anonymize presidioapi.AnonymizeInterface) TransformerExecutorOption {
	return func(c *TransformerExecutorConfig) {
		c.transformPiiText = &transformPiiTextConfig{
			analyze:   analyze,
			anonymize: anonymize,
		}
	}
}

func InitializeTransformer(transformerMapping *mgmtv1alpha1.JobMappingTransformer, opts ...TransformerExecutorOption) (*TransformerExecutor, error) {
	return InitializeTransformerByConfigType(transformerMapping.GetConfig(), opts...)
}

func InitializeTransformerByConfigType(transformerConfig *mgmtv1alpha1.TransformerConfig, opts ...TransformerExecutorOption) (*TransformerExecutor, error) {
	execCfg := &TransformerExecutorConfig{}
	for _, opt := range opts {
		opt(execCfg)
	}

	maxLength := int64(10000) // TODO: update this based on colInfo if available
	switch transformerConfig.GetConfig().(type) {
	case *mgmtv1alpha1.TransformerConfig_PassthroughConfig:
		return &TransformerExecutor{
			Opts: nil,
			Mutate: func(value any, opts any) (any, error) {
				return value, nil
			},
		}, nil
	case *mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig:
		categories := transformerConfig.GetGenerateCategoricalConfig().GetCategories()
		opts, err := NewGenerateCategoricalOpts(categories, nil)
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

	case *mgmtv1alpha1.TransformerConfig_GenerateBoolConfig:
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

	case *mgmtv1alpha1.TransformerConfig_TransformStringConfig:
		pl := transformerConfig.GetTransformStringConfig().GetPreserveLength()
		minLength := int64(3) // TODO: pull this value from the database schema
		opts, err := NewTransformStringOpts(&pl, &minLength, &maxLength, nil)
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
	case *mgmtv1alpha1.TransformerConfig_TransformInt64Config:
		rMin := transformerConfig.GetTransformInt64Config().GetRandomizationRangeMin()
		rMax := transformerConfig.GetTransformInt64Config().GetRandomizationRangeMax()
		opts, err := NewTransformInt64Opts(rMin, rMax, nil)
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

	case *mgmtv1alpha1.TransformerConfig_TransformFullNameConfig:
		pl := transformerConfig.GetTransformFullNameConfig().GetPreserveLength()
		opts, err := NewTransformFullNameOpts(&maxLength, &pl, nil)
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

	case *mgmtv1alpha1.TransformerConfig_GenerateEmailConfig:
		config := transformerConfig.GetGenerateEmailConfig()
		var emailType *string
		if config.EmailType != nil {
			emailTypeStr := config.GetEmailType().String()
			emailType = &emailTypeStr
		}
		opts, err := NewGenerateEmailOpts(&maxLength, emailType, nil)
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

	case *mgmtv1alpha1.TransformerConfig_TransformEmailConfig:
		config := transformerConfig.GetTransformEmailConfig()
		var emailType *string
		if config.EmailType != nil {
			emailTypeStr := config.GetEmailType().String()
			emailType = &emailTypeStr
		}
		var invalidEmailAction *string
		if config.InvalidEmailAction != nil {
			invalidEmailActionStr := config.GetInvalidEmailAction().String()
			invalidEmailAction = &invalidEmailActionStr
		}
		var excludedDomains any = config.GetExcludedDomains()
		opts, err := NewTransformEmailOpts(
			&config.PreserveDomain,
			&config.PreserveLength,
			&excludedDomains,
			&maxLength,
			nil,
			emailType,
			invalidEmailAction,
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

	case *mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig:
		luhn := transformerConfig.GetGenerateCardNumberConfig().GetValidLuhn()
		opts, err := NewGenerateCardNumberOpts(luhn, nil)
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

	case *mgmtv1alpha1.TransformerConfig_GenerateCityConfig:
		opts, err := NewGenerateCityOpts(maxLength, nil)
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

	case *mgmtv1alpha1.TransformerConfig_GenerateE164PhoneNumberConfig:
		minValue := transformerConfig.GetGenerateE164PhoneNumberConfig().GetMin()
		maxValue := transformerConfig.GetGenerateE164PhoneNumberConfig().GetMax()
		opts, err := NewGenerateInternationalPhoneNumberOpts(minValue, maxValue, nil)
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
	case *mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig:
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

	case *mgmtv1alpha1.TransformerConfig_GenerateFloat64Config:
		config := transformerConfig.GetGenerateFloat64Config()
		opts, err := NewGenerateFloat64Opts(
			&config.RandomizeSign,
			config.GetMin(),
			config.GetMax(),
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

	case *mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig:
		opts, err := NewGenerateFullAddressOpts(maxLength, nil)
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

	case *mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig:
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

	case *mgmtv1alpha1.TransformerConfig_GenerateGenderConfig:
		ab := transformerConfig.GetGenerateGenderConfig().GetAbbreviate()
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

	case *mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneNumberConfig:
		opts, err := NewGenerateInt64PhoneNumberOpts(nil)
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

	case *mgmtv1alpha1.TransformerConfig_GenerateInt64Config:
		config := transformerConfig.GetGenerateInt64Config()
		opts, err := NewGenerateInt64Opts(&config.RandomizeSign, config.GetMin(), config.GetMax(), nil)
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

	case *mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig:
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

	case *mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig:
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

	case *mgmtv1alpha1.TransformerConfig_GenerateSsnConfig:
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

	case *mgmtv1alpha1.TransformerConfig_GenerateStateConfig:
		generateFullName := transformerConfig.GetGenerateStateConfig().GetGenerateFullName()
		opts, err := NewGenerateStateOpts(&generateFullName, nil)
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

	case *mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig:
		opts, err := NewGenerateStreetAddressOpts(maxLength, nil)
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

	case *mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig:
		minValue := transformerConfig.GetGenerateStringPhoneNumberConfig().GetMin()
		maxValue := transformerConfig.GetGenerateStringPhoneNumberConfig().GetMax()
		minValue = transformer_utils.MinInt(minValue, maxLength)
		maxValue = transformer_utils.Ceil(maxValue, maxLength)
		opts, err := NewGenerateStringPhoneNumberOpts(minValue, maxValue, nil)
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

	case *mgmtv1alpha1.TransformerConfig_GenerateStringConfig:
		config := transformerConfig.GetGenerateStringConfig()
		opts, err := NewGenerateRandomStringOpts(config.GetMin(), config.GetMax(), nil)
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

	case *mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig:
		opts, err := NewGenerateUnixTimestampOpts(nil)
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

	case *mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig:
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

	case *mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig:
		opts, err := NewGenerateUTCTimestampOpts(nil)
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

	case *mgmtv1alpha1.TransformerConfig_GenerateUuidConfig:
		ih := transformerConfig.GetGenerateUuidConfig().GetIncludeHyphens()
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

	case *mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig:
		opts, err := NewGenerateZipcodeOpts(nil)
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

	case *mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig:
		config := transformerConfig.GetTransformE164PhoneNumberConfig()
		opts, err := NewTransformE164PhoneNumberOpts(config.GetPreserveLength(), &maxLength, nil)
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

	case *mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig:
		config := transformerConfig.GetTransformFirstNameConfig()
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

	case *mgmtv1alpha1.TransformerConfig_TransformFloat64Config:
		config := transformerConfig.GetTransformFloat64Config()
		opts, err := NewTransformFloat64Opts(
			config.GetRandomizationRangeMin(),
			config.GetRandomizationRangeMax(),
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

	case *mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig:
		config := transformerConfig.GetTransformInt64PhoneNumberConfig()
		opts, err := NewTransformInt64PhoneNumberOpts(config.GetPreserveLength(), nil)
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

	case *mgmtv1alpha1.TransformerConfig_TransformLastNameConfig:
		config := transformerConfig.GetTransformLastNameConfig()
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

	case *mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig:
		config := transformerConfig.GetTransformPhoneNumberConfig()
		opts, err := NewTransformStringPhoneNumberOpts(config.GetPreserveLength(), maxLength, nil)
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

	case *mgmtv1alpha1.TransformerConfig_Nullconfig:
		return &TransformerExecutor{
			Opts: nil,
			Mutate: func(value any, opts any) (any, error) {
				return "null", nil
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig:
		return &TransformerExecutor{
			Opts: nil,
			Mutate: func(value any, opts any) (any, error) {
				return `"DEFAULT"`, nil
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig:
		config := transformerConfig.GetTransformCharacterScrambleConfig()
		opts, err := NewTransformCharacterScrambleOpts(config.UserProvidedRegex, nil)
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

	case *mgmtv1alpha1.TransformerConfig_GenerateCountryConfig:
		generateFullName := transformerConfig.GetGenerateCountryConfig().GenerateFullName
		opts, err := NewGenerateCountryOpts(&generateFullName, nil)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateCountry().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig:
		if execCfg.transformPiiText == nil {
			return nil, fmt.Errorf("transformer: TransformPiiText is not enabled: %w", errors.ErrUnsupported)
		}
		config := transformerConfig.GetTransformPiiTextConfig()

		return &TransformerExecutor{
			Opts: nil,
			Mutate: func(value, opts any) (any, error) {
				valueStr, ok := value.(string)
				if !ok {
					return nil, fmt.Errorf("expected value to be of type string. %T", value)
				}
				return ee_transformer_fns.TransformPiiText(
					context.Background(),
					execCfg.transformPiiText.analyze, execCfg.transformPiiText.anonymize,
					config, valueStr,
				)
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported transformer: %v", transformerConfig)
	}
}
