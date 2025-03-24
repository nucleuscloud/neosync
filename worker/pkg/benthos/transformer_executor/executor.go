package transformer_executor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/dop251/goja"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	ee_transformer_fns "github.com/nucleuscloud/neosync/internal/ee/transformers/functions"
	"github.com/nucleuscloud/neosync/internal/javascript"
	javascript_userland "github.com/nucleuscloud/neosync/internal/javascript/userland"
	"github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
)

type TransformerExecutor struct {
	Opts   any
	Mutate func(value any, opts any) (any, error)
}

type TransformerExecutorOption func(c *TransformerExecutorConfig)

type TransformerExecutorConfig struct {
	transformPiiText               *transformPiiTextConfig
	userDefinedTransformerResolver UserDefinedTransformerResolver
	logger                         *slog.Logger
}

type transformPiiTextConfig struct {
	analyze   presidioapi.AnalyzeInterface
	anonymize presidioapi.AnonymizeInterface

	neosyncOperatorApi ee_transformer_fns.NeosyncOperatorApi

	defaultLanguage *string
}

func WithTransformPiiTextConfig(analyze presidioapi.AnalyzeInterface, anonymize presidioapi.AnonymizeInterface, neosyncOperatorApi ee_transformer_fns.NeosyncOperatorApi, defaultLanguage *string) TransformerExecutorOption {
	return func(c *TransformerExecutorConfig) {
		c.transformPiiText = &transformPiiTextConfig{
			analyze:            analyze,
			anonymize:          anonymize,
			neosyncOperatorApi: neosyncOperatorApi,
			defaultLanguage:    defaultLanguage,
		}
	}
}

func WithLogger(logger *slog.Logger) TransformerExecutorOption {
	return func(c *TransformerExecutorConfig) {
		c.logger = logger
	}
}

func InitializeTransformer(transformerMapping *mgmtv1alpha1.JobMappingTransformer, opts ...TransformerExecutorOption) (*TransformerExecutor, error) {
	return InitializeTransformerByConfigType(transformerMapping.GetConfig(), opts...)
}

type UserDefinedTransformerResolver interface {
	GetUserDefinedTransformer(ctx context.Context, id string) (*mgmtv1alpha1.TransformerConfig, error)
}

func WithUserDefinedTransformerResolver(resolver UserDefinedTransformerResolver) TransformerExecutorOption {
	return func(c *TransformerExecutorConfig) {
		c.userDefinedTransformerResolver = resolver
	}
}

func InitializeTransformerByConfigType(transformerConfig *mgmtv1alpha1.TransformerConfig, opts ...TransformerExecutorOption) (*TransformerExecutor, error) {
	execCfg := &TransformerExecutorConfig{logger: slog.Default()}
	for _, opt := range opts {
		opt(execCfg)
	}

	maxLength := int64(100) // TODO: update this based on colInfo if available
	switch typedCfg := transformerConfig.GetConfig().(type) {
	case *mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig:
		if execCfg.userDefinedTransformerResolver == nil {
			return nil, fmt.Errorf("user defined transformer resolver is not set")
		}
		config := typedCfg.UserDefinedTransformerConfig
		if config == nil {
			return nil, fmt.Errorf("user defined transformer config is nil")
		}
		resolvedConfig, err := execCfg.userDefinedTransformerResolver.GetUserDefinedTransformer(context.Background(), config.GetId())
		if err != nil {
			return nil, err
		}
		return InitializeTransformerByConfigType(resolvedConfig, opts...)
	case *mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig:
		config := typedCfg.GenerateJavascriptConfig
		if config == nil {
			return nil, fmt.Errorf("generate javascript config is nil")
		}

		valueApi := newAnonValueApi()
		runner, err := javascript.NewDefaultValueRunner(valueApi, execCfg.logger)
		if err != nil {
			return nil, err
		}
		jsCode, propertyPath := javascript_userland.GetSingleGenerateFunction(config.GetCode())
		program, err := goja.Compile("main.js", jsCode, false)
		if err != nil {
			return nil, err
		}

		return &TransformerExecutor{
			Opts: nil,
			Mutate: func(value any, opts any) (any, error) {
				inputMessage, err := NewMessage(map[string]any{})
				if err != nil {
					return nil, fmt.Errorf("failed to create input message: %w", err)
				}
				valueApi.SetMessage(inputMessage)
				_, err = runner.Run(context.Background(), program)
				if err != nil {
					return nil, fmt.Errorf("failed to run program: %w", err)
				}
				updatedValue, err := valueApi.GetPropertyPathValue(propertyPath)
				if err != nil {
					return nil, fmt.Errorf("failed to get property path value: %w", err)
				}
				return updatedValue, nil
			},
		}, nil
	case *mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig:
		config := typedCfg.TransformJavascriptConfig
		if config == nil {
			return nil, fmt.Errorf("transform javascript config is nil")
		}

		valueApi := newAnonValueApi()
		runner, err := javascript.NewDefaultValueRunner(valueApi, execCfg.logger)
		if err != nil {
			return nil, err
		}
		jsCode, propertyPath := javascript_userland.GetSingleTransformFunction(config.GetCode())
		program, err := goja.Compile("main.js", jsCode, false)
		if err != nil {
			return nil, err
		}

		return &TransformerExecutor{
			Opts: nil,
			Mutate: func(value any, opts any) (any, error) {
				inputMessage, err := NewMessage(map[string]any{
					propertyPath: value,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to create input message: %w", err)
				}
				valueApi.SetMessage(inputMessage)
				_, err = runner.Run(context.Background(), program)
				if err != nil {
					return nil, fmt.Errorf("failed to run program: %w", err)
				}
				updatedValue, err := valueApi.GetPropertyPathValue(propertyPath)
				if err != nil {
					return nil, fmt.Errorf("failed to get property path value: %w", err)
				}
				return updatedValue, nil
			},
		}, nil
	case *mgmtv1alpha1.TransformerConfig_PassthroughConfig:
		return &TransformerExecutor{
			Opts: nil,
			Mutate: func(value any, opts any) (any, error) {
				return value, nil
			},
		}, nil
	case *mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig:
		config := transformerConfig.GetGenerateCategoricalConfig()
		opts, err := transformers.NewGenerateCategoricalOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateCategorical().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateBoolConfig:
		config := transformerConfig.GetGenerateBoolConfig()
		opts, err := transformers.NewGenerateBoolOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateBool().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_TransformStringConfig:
		config := transformerConfig.GetTransformStringConfig()
		minLength := int64(3) // TODO: pull this value from the database schema
		opts, err := transformers.NewTransformStringOptsFromConfig(config, &minLength, &maxLength)
		if err != nil {
			return nil, err
		}
		transform := transformers.NewTransformString().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil
	case *mgmtv1alpha1.TransformerConfig_TransformInt64Config:
		config := transformerConfig.GetTransformInt64Config()
		opts, err := transformers.NewTransformInt64OptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		transform := transformers.NewTransformInt64().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_TransformFullNameConfig:
		config := transformerConfig.GetTransformFullNameConfig()
		opts, err := transformers.NewTransformFullNameOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		transform := transformers.NewTransformFullName().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateEmailConfig:
		config := transformerConfig.GetGenerateEmailConfig()
		opts, err := transformers.NewGenerateEmailOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateEmail().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_TransformEmailConfig:
		config := transformerConfig.GetTransformEmailConfig()
		opts, err := transformers.NewTransformEmailOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		transform := transformers.NewTransformEmail().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig:
		config := transformerConfig.GetGenerateCardNumberConfig()
		opts, err := transformers.NewGenerateCardNumberOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateCardNumber().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateCityConfig:
		config := transformerConfig.GetGenerateCityConfig()
		opts, err := transformers.NewGenerateCityOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateCity().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateE164PhoneNumberConfig:
		config := transformerConfig.GetGenerateE164PhoneNumberConfig()
		opts, err := transformers.NewGenerateInternationalPhoneNumberOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateInternationalPhoneNumber().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil
	case *mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig:
		config := transformerConfig.GetGenerateFirstNameConfig()
		opts, err := transformers.NewGenerateFirstNameOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateFirstName().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateFloat64Config:
		config := transformerConfig.GetGenerateFloat64Config()
		opts, err := transformers.NewGenerateFloat64OptsFromConfig(config, nil)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateFloat64().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig:
		config := transformerConfig.GetGenerateFullAddressConfig()
		opts, err := transformers.NewGenerateFullAddressOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateFullAddress().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig:
		config := transformerConfig.GetGenerateFullNameConfig()
		opts, err := transformers.NewGenerateFullNameOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateFullName().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateGenderConfig:
		config := transformerConfig.GetGenerateGenderConfig()
		opts, err := transformers.NewGenerateGenderOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateGender().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneNumberConfig:
		config := transformerConfig.GetGenerateInt64PhoneNumberConfig()
		opts, err := transformers.NewGenerateInt64PhoneNumberOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateInt64PhoneNumber().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateInt64Config:
		config := transformerConfig.GetGenerateInt64Config()
		opts, err := transformers.NewGenerateInt64OptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateInt64().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig:
		config := transformerConfig.GetGenerateLastNameConfig()
		opts, err := transformers.NewGenerateLastNameOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateLastName().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig:
		config := transformerConfig.GetGenerateSha256HashConfig()
		opts, err := transformers.NewGenerateSHA256HashOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateSHA256Hash().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateSsnConfig:
		config := transformerConfig.GetGenerateSsnConfig()
		opts, err := transformers.NewGenerateSSNOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateSSN().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateStateConfig:
		config := transformerConfig.GetGenerateStateConfig()
		opts, err := transformers.NewGenerateStateOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateState().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig:
		config := transformerConfig.GetGenerateStreetAddressConfig()
		opts, err := transformers.NewGenerateStreetAddressOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateStreetAddress().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig:
		config := transformerConfig.GetGenerateStringPhoneNumberConfig()
		opts, err := transformers.NewGenerateStringPhoneNumberOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateStringPhoneNumber().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateStringConfig:
		config := transformerConfig.GetGenerateStringConfig()
		opts, err := transformers.NewGenerateRandomStringOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateRandomString().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig:
		config := transformerConfig.GetGenerateUnixtimestampConfig()
		opts, err := transformers.NewGenerateUnixTimestampOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateUnixTimestamp().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig:
		config := transformerConfig.GetGenerateUsernameConfig()
		opts, err := transformers.NewGenerateUsernameOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateUsername().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig:
		config := transformerConfig.GetGenerateUtctimestampConfig()
		opts, err := transformers.NewGenerateUTCTimestampOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateUTCTimestamp().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateUuidConfig:
		config := transformerConfig.GetGenerateUuidConfig()
		opts, err := transformers.NewGenerateUUIDOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateUUID().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig:
		config := transformerConfig.GetGenerateZipcodeConfig()
		opts, err := transformers.NewGenerateZipcodeOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateZipcode().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig:
		config := transformerConfig.GetTransformE164PhoneNumberConfig()
		opts, err := transformers.NewTransformE164PhoneNumberOptsFromConfig(config, nil)
		if err != nil {
			return nil, err
		}
		transform := transformers.NewTransformE164PhoneNumber().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig:
		config := transformerConfig.GetTransformFirstNameConfig()
		opts, err := transformers.NewTransformFirstNameOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		transform := transformers.NewTransformFirstName().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_TransformFloat64Config:
		config := transformerConfig.GetTransformFloat64Config()
		opts, err := transformers.NewTransformFloat64OptsFromConfig(config, nil, nil)
		if err != nil {
			return nil, err
		}
		transform := transformers.NewTransformFloat64().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig:
		config := transformerConfig.GetTransformInt64PhoneNumberConfig()
		opts, err := transformers.NewTransformInt64PhoneNumberOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		transform := transformers.NewTransformInt64PhoneNumber().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_TransformLastNameConfig:
		config := transformerConfig.GetTransformLastNameConfig()
		opts, err := transformers.NewTransformLastNameOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		transform := transformers.NewTransformLastName().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig:
		config := transformerConfig.GetTransformPhoneNumberConfig()
		opts, err := transformers.NewTransformStringPhoneNumberOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		transform := transformers.NewTransformStringPhoneNumber().Transform
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
		opts, err := transformers.NewTransformCharacterScrambleOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		transform := transformers.NewTransformCharacterScramble().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateCountryConfig:
		config := transformerConfig.GetGenerateCountryConfig()
		opts, err := transformers.NewGenerateCountryOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateCountry().Generate
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
		if config == nil {
			config = &mgmtv1alpha1.TransformPiiText{}
		}
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
					execCfg.transformPiiText.analyze, execCfg.transformPiiText.anonymize, execCfg.transformPiiText.neosyncOperatorApi,
					config,
					valueStr,
					execCfg.logger,
				)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateBusinessNameConfig:
		config := transformerConfig.GetGenerateBusinessNameConfig()
		opts, err := transformers.NewGenerateBusinessNameOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateBusinessName().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_GenerateIpAddressConfig:
		config := transformerConfig.GetGenerateIpAddressConfig()
		opts, err := transformers.NewGenerateIpAddressOptsFromConfig(config, &maxLength)
		if err != nil {
			return nil, err
		}
		generate := transformers.NewGenerateIpAddress().Generate
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil
	case *mgmtv1alpha1.TransformerConfig_TransformUuidConfig:
		config := transformerConfig.GetTransformUuidConfig()
		opts, err := transformers.NewTransformUuidOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		transform := transformers.NewTransformUuid().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	case *mgmtv1alpha1.TransformerConfig_ScrambleIdentityConfig:
		config := transformerConfig.GetScrambleIdentityConfig()
		opts, err := transformers.NewTransformIdentityScrambleOptsFromConfig(config)
		if err != nil {
			return nil, err
		}
		transform := transformers.NewTransformIdentityScramble().Transform
		return &TransformerExecutor{
			Opts: opts,
			Mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported transformer: %T", typedCfg)
	}
}
