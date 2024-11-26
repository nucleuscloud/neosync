package transformers

import (
	"context"
	"errors"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	ee_transformer_fns "github.com/nucleuscloud/neosync/internal/ee/transformers/functions"
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

	defaultLanguage *string
}

func WithTransformPiiTextConfig(analyze presidioapi.AnalyzeInterface, anonymize presidioapi.AnonymizeInterface, defaultLanguage *string) TransformerExecutorOption {
	return func(c *TransformerExecutorConfig) {
		c.transformPiiText = &transformPiiTextConfig{
			analyze:         analyze,
			anonymize:       anonymize,
			defaultLanguage: defaultLanguage,
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

	maxLength := int64(100) // TODO: update this based on colInfo if available
	switch transformerConfig.GetConfig().(type) {
	case *mgmtv1alpha1.TransformerConfig_PassthroughConfig:
		return &TransformerExecutor{
			Opts: nil,
			Mutate: func(value any, opts any) (any, error) {
				return value, nil
			},
		}, nil
	case *mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig:
		config := transformerConfig.GetGenerateCategoricalConfig()
		opts, err := NewGenerateCategoricalOptsFromConfig(config)
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
		config := transformerConfig.GetGenerateBoolConfig()
		opts, err := NewGenerateBoolOptsFromConfig(config)
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
		config := transformerConfig.GetTransformStringConfig()
		minLength := int64(3) // TODO: pull this value from the database schema
		opts, err := NewTransformStringOptsFromConfig(config, &minLength, &maxLength)
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
		config := transformerConfig.GetTransformInt64Config()
		opts, err := NewTransformInt64OptsFromConfig(config)
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
		config := transformerConfig.GetTransformFullNameConfig()
		opts, err := NewTransformFullNameOptsFromConfig(config, &maxLength)
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
		opts, err := NewGenerateEmailOptsFromConfig(config, &maxLength)
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
		opts, err := NewTransformEmailOptsFromConfig(config, &maxLength)
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
		config := transformerConfig.GetGenerateCardNumberConfig()
		opts, err := NewGenerateCardNumberOptsFromConfig(config)
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
		config := transformerConfig.GetGenerateCityConfig()
		opts, err := NewGenerateCityOptsFromConfig(config, &maxLength)
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
		config := transformerConfig.GetGenerateE164PhoneNumberConfig()
		opts, err := NewGenerateInternationalPhoneNumberOptsFromConfig(config)
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
		config := transformerConfig.GetGenerateFirstNameConfig()
		opts, err := NewGenerateFirstNameOptsFromConfig(config, &maxLength)
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
		opts, err := NewGenerateFloat64OptsFromConfig(config, nil)
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
		config := transformerConfig.GetGenerateFullAddressConfig()
		opts, err := NewGenerateFullAddressOptsFromConfig(config, &maxLength)
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
		config := transformerConfig.GetGenerateFullNameConfig()
		opts, err := NewGenerateFullNameOptsFromConfig(config, &maxLength)
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
		config := transformerConfig.GetGenerateGenderConfig()
		opts, err := NewGenerateGenderOptsFromConfig(config, &maxLength)
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
		config := transformerConfig.GetGenerateInt64PhoneNumberConfig()
		opts, err := NewGenerateInt64PhoneNumberOptsFromConfig(config)
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
		opts, err := NewGenerateInt64OptsFromConfig(config)
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
		config := transformerConfig.GetGenerateLastNameConfig()
		opts, err := NewGenerateLastNameOptsFromConfig(config, &maxLength)
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
		config := transformerConfig.GetGenerateSha256HashConfig()
		opts, err := NewGenerateSHA256HashOptsFromConfig(config)
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
		config := transformerConfig.GetGenerateSsnConfig()
		opts, err := NewGenerateSSNOptsFromConfig(config)
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
		config := transformerConfig.GetGenerateStateConfig()
		opts, err := NewGenerateStateOptsFromConfig(config)
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
		config := transformerConfig.GetGenerateStreetAddressConfig()
		opts, err := NewGenerateStreetAddressOptsFromConfig(config, &maxLength)
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
		config := transformerConfig.GetGenerateStringPhoneNumberConfig()
		opts, err := NewGenerateStringPhoneNumberOptsFromConfig(config)
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
		opts, err := NewGenerateRandomStringOptsFromConfig(config, &maxLength)
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
		config := transformerConfig.GetGenerateUnixtimestampConfig()
		opts, err := NewGenerateUnixTimestampOptsFromConfig(config)
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
		config := transformerConfig.GetGenerateUsernameConfig()
		opts, err := NewGenerateUsernameOptsFromConfig(config, &maxLength)
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
		config := transformerConfig.GetGenerateUtctimestampConfig()
		opts, err := NewGenerateUTCTimestampOptsFromConfig(config)
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
		config := transformerConfig.GetGenerateUuidConfig()
		opts, err := NewGenerateUUIDOptsFromConfig(config)
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
		config := transformerConfig.GetGenerateZipcodeConfig()
		opts, err := NewGenerateZipcodeOptsFromConfig(config)
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
		opts, err := NewTransformE164PhoneNumberOptsFromConfig(config, nil)
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
		opts, err := NewTransformFirstNameOptsFromConfig(config, &maxLength)
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
		opts, err := NewTransformFloat64OptsFromConfig(config, nil, nil)
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
		opts, err := NewTransformInt64PhoneNumberOptsFromConfig(config)
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
		opts, err := NewTransformLastNameOptsFromConfig(config, &maxLength)
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
		opts, err := NewTransformStringPhoneNumberOptsFromConfig(config, &maxLength)
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
		opts, err := NewTransformCharacterScrambleOptsFromConfig(config)
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
		config := transformerConfig.GetGenerateCountryConfig()
		opts, err := NewGenerateCountryOptsFromConfig(config)
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
		if config.GetLanguage() == "" && execCfg.transformPiiText.defaultLanguage != nil {
			config.Language = execCfg.transformPiiText.defaultLanguage
		}

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
					config,
					valueStr,
				)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateBusinessNameConfig:
		config := transformerConfig.GetGenerateBusinessNameConfig()
		opts, err := NewGenerateBusinessNameOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateBusinessName().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateIpAddressConfig:
		config := transformerConfig.GetGenerateIpAddressConfig()
		opts, err := NewGenerateIpAddressOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		generate := NewGenerateIpAddress().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported transformer: %v", transformerConfig)
	}
}
