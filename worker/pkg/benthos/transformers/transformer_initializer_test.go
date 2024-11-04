package transformers

import (
	"strconv"
	"testing"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_InitializeTransformerByConfigType(t *testing.T) {
	t.Run("PassthroughConfig", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("test", nil)
		require.NoError(t, err)
		require.Equal(t, "test", result)
	})

	t.Run("GenerateCategoricalConfig", func(t *testing.T) {
		categories := "A,B,C"
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig{
				GenerateCategoricalConfig: &mgmtv1alpha1.GenerateCategorical{
					Categories: &categories,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.Contains(t, []string{"A", "B", "C"}, result)
	})

	t.Run("GenerateCategoricalConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig{
				GenerateCategoricalConfig: &mgmtv1alpha1.GenerateCategorical{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateCategoricalConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateBoolConfig", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, bool(true), result)
	})

	t.Run("GenerateBoolConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, bool(true), result)
	})

	t.Run("TransformStringConfig", func(t *testing.T) {
		preserveLength := true
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
				TransformStringConfig: &mgmtv1alpha1.TransformString{
					PreserveLength: &preserveLength,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("test", executor.Opts)
		require.NoError(t, err)
		require.Len(t, *result.(*string), 4)
	})

	t.Run("TransformStringConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
				TransformStringConfig: &mgmtv1alpha1.TransformString{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("test", executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("TransformStringConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("test", executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("TransformInt64Config", func(t *testing.T) {
		rmin, rmax := int64(1), int64(5)
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{
				TransformInt64Config: &mgmtv1alpha1.TransformInt64{
					RandomizationRangeMin: &rmin,
					RandomizationRangeMax: &rmax,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(int64(50), executor.Opts)
		require.NoError(t, err)
		require.GreaterOrEqual(t, *result.(*int64), int64(48))
		require.LessOrEqual(t, *result.(*int64), int64(55))
	})

	t.Run("TransformInt64Config_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{
				TransformInt64Config: &mgmtv1alpha1.TransformInt64{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(int64(50), executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("TransformInt64Config_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(int64(50), executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("TransformFullNameConfig", func(t *testing.T) {
		preserveLength := true
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
				TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{
					PreserveLength: &preserveLength,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("John Doe", executor.Opts)
		require.NoError(t, err)
		require.Len(t, *result.(*string), 8)
	})
	t.Run("TransformFullNameConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
				TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("John Doe", executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("TransformFullNameConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("John Doe", executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateEmailConfig", func(t *testing.T) {
		emailType := mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_FULLNAME
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
				GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{
					EmailType: &emailType,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateEmailConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
				GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateEmailConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("TransformEmailConfig", func(t *testing.T) {
		preserve := true
		emailType := mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_FULLNAME
		invalidEmailAction := mgmtv1alpha1.InvalidEmailAction_INVALID_EMAIL_ACTION_GENERATE
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
				TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
					PreserveDomain:     &preserve,
					PreserveLength:     &preserve,
					EmailType:          &emailType,
					InvalidEmailAction: &invalidEmailAction,
					ExcludedDomains:    []string{"gmail", "yahoo"},
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("test@example.com", executor.Opts)
		require.NoError(t, err)
		require.Regexp(t, `^[a-zA-Z0-9._%+-]+@example\.com$`, *result.(*string))
		require.Len(t, *result.(*string), 16)
	})

	t.Run("TransformEmailConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
				TransformEmailConfig: &mgmtv1alpha1.TransformEmail{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("test@example.com", executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})
	t.Run("TransformEmailConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
				TransformEmailConfig: &mgmtv1alpha1.TransformEmail{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("test@example.com", executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateCardNumberConfig", func(t *testing.T) {
		valid := true
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{
				GenerateCardNumberConfig: &mgmtv1alpha1.GenerateCardNumber{
					ValidLuhn: &valid,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.Regexp(t, `^\d{13,19}$`, result)
	})

	t.Run("GenerateCardNumber_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{
				GenerateCardNumberConfig: &mgmtv1alpha1.GenerateCardNumber{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.Regexp(t, `^\d{13,19}$`, result)
	})

	t.Run("GenerateCardNumber_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{
				GenerateCardNumberConfig: &mgmtv1alpha1.GenerateCardNumber{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.Regexp(t, `^\d{13,19}$`, result)
	})

	t.Run("GenerateCityConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
				GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateCityConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateInternationalPhoneNumberConfig", func(t *testing.T) {
		rmin, rmax := int64(10), int64(10)
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateE164PhoneNumberConfig{
				GenerateE164PhoneNumberConfig: &mgmtv1alpha1.GenerateE164PhoneNumber{
					Min: &rmin,
					Max: &rmax,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.Regexp(t, `^\+\d{10}$`, result)
	})

	t.Run("GenerateInternationPhoneNumberConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateE164PhoneNumberConfig{
				GenerateE164PhoneNumberConfig: &mgmtv1alpha1.GenerateE164PhoneNumber{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateInternationPhoneNumberConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateE164PhoneNumberConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateFirstNameConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
				GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateFirstNameConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateFloat64Config", func(t *testing.T) {
		randomizeSign := true
		rmin, rmax := float64(-10), float64(10)
		precision := int64(2)
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{
				GenerateFloat64Config: &mgmtv1alpha1.GenerateFloat64{
					RandomizeSign: &randomizeSign,
					Min:           &rmin,
					Max:           &rmax,
					Precision:     &precision,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, float64(0), result)
		floatResult := result.(float64)
		require.GreaterOrEqual(t, floatResult, rmin)
		require.LessOrEqual(t, floatResult, rmax)
	})

	t.Run("GenerateFloat64Config_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{
				GenerateFloat64Config: &mgmtv1alpha1.GenerateFloat64{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, float64(0), result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateFloat64Config_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, float64(0), result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateFullAddressConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig{
				GenerateFullAddressConfig: &mgmtv1alpha1.GenerateFullAddress{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateFullAddressConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateFullNameConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
				GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateFullNameConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateGenderConfig", func(t *testing.T) {
		abb := true
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateGenderConfig{
				GenerateGenderConfig: &mgmtv1alpha1.GenerateGender{
					Abbreviate: &abb,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
		require.Len(t, result, 1)
	})

	t.Run("GenerateGenderConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateGenderConfig{
				GenerateGenderConfig: &mgmtv1alpha1.GenerateGender{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateGenderConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateGenderConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateInt64PhoneNumberConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneNumberConfig{
				GenerateInt64PhoneNumberConfig: &mgmtv1alpha1.GenerateInt64PhoneNumber{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, int64(0), result)
		require.Regexp(t, `^[1-9]\d{9}$`, strconv.FormatInt(result.(int64), 10))
	})

	t.Run("GenerateInt64PhoneNumberConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneNumberConfig{
				GenerateInt64PhoneNumberConfig: &mgmtv1alpha1.GenerateInt64PhoneNumber{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, int64(0), result)
		require.Regexp(t, `^[1-9]\d{9}$`, strconv.FormatInt(result.(int64), 10))
	})

	t.Run("GenerateInt64Config", func(t *testing.T) {
		randomizeSign := true
		rmin, rmax := int64(-1000), int64(1000)
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
				GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
					RandomizeSign: &randomizeSign,
					Min:           &rmin,
					Max:           &rmax,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, int64(0), result)
		intResult := result.(int64)
		require.GreaterOrEqual(t, intResult, rmin)
		require.LessOrEqual(t, intResult, rmax)
	})

	t.Run("GenerateInt64Config_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
				GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, int64(0), result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateInt64Config_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, int64(0), result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateLastNameConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig{
				GenerateLastNameConfig: &mgmtv1alpha1.GenerateLastName{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
		require.Regexp(t, `^[A-Z][a-z]+$`, result)
	})

	t.Run("GenerateLastNameConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
		require.Regexp(t, `^[A-Z][a-z]+$`, result)
	})

	t.Run("GenerateSha256HashConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig{
				GenerateSha256HashConfig: &mgmtv1alpha1.GenerateSha256Hash{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.Regexp(t, `^[a-f0-9]{64}$`, result)
	})

	t.Run("GenerateSha256HashConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.Regexp(t, `^[a-f0-9]{64}$`, result)
	})

	t.Run("GenerateSsnConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
				GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.Regexp(t, `^\d{3}-\d{2}-\d{4}$`, result)
	})

	t.Run("GenerateSsnConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.Regexp(t, `^\d{3}-\d{2}-\d{4}$`, result)
	})

	t.Run("GenerateStateConfig", func(t *testing.T) {
		genFullName := true
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStateConfig{
				GenerateStateConfig: &mgmtv1alpha1.GenerateState{
					GenerateFullName: &genFullName,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
		require.Regexp(t, `^[A-Z][a-z]+( [A-Z][a-z]+)*$`, result)
	})

	t.Run("GenerateStateConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStateConfig{
				GenerateStateConfig: &mgmtv1alpha1.GenerateState{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
		require.Len(t, result, 2)
	})

	t.Run("GenerateStateConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStateConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
		require.Len(t, result, 2)
	})

	t.Run("GenerateStreetAddressConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig{
				GenerateStreetAddressConfig: &mgmtv1alpha1.GenerateStreetAddress{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateStreetAddressConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateStringPhoneNumberConfig", func(t *testing.T) {
		rmin, rmax := int64(10), int64(12)
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig{
				GenerateStringPhoneNumberConfig: &mgmtv1alpha1.GenerateStringPhoneNumber{
					Min: &rmin,
					Max: &rmax,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.Regexp(t, `^\+?1?\d{10,12}$`, result)
	})

	t.Run("GenerateStringPhoneNumberConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig{
				GenerateStringPhoneNumberConfig: &mgmtv1alpha1.GenerateStringPhoneNumber{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateStringPhoneNumberConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateStringConfig", func(t *testing.T) {
		rmin, rmax := int64(5), int64(10)
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
				GenerateStringConfig: &mgmtv1alpha1.GenerateString{
					Min: &rmin,
					Max: &rmax,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.GreaterOrEqual(t, len(result.(string)), int(rmin))
		require.LessOrEqual(t, len(result.(string)), int(rmax))
	})

	t.Run("GenerateStringConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
				GenerateStringConfig: &mgmtv1alpha1.GenerateString{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateStringConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateUnixtimestampConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{
				GenerateUnixtimestampConfig: &mgmtv1alpha1.GenerateUnixTimestamp{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, int64(0), result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateUnixtimestampConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, int64(0), result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateUsernameConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig{
				GenerateUsernameConfig: &mgmtv1alpha1.GenerateUsername{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateUsernameConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateUtctimestampConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig{
				GenerateUtctimestampConfig: &mgmtv1alpha1.GenerateUtcTimestamp{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, time.Now(), result)
	})

	t.Run("GenerateUtctimestampConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, time.Now(), result)
	})

	t.Run("GenerateUuidConfig", func(t *testing.T) {
		hyphens := false
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
				GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
					IncludeHyphens: &hyphens,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateUuidConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
				GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateUuidConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateZipcodeConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig{
				GenerateZipcodeConfig: &mgmtv1alpha1.GenerateZipcode{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateZipcodeConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("TransformE164PhoneNumberConfig", func(t *testing.T) {
		preserveLength := true
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig{
				TransformE164PhoneNumberConfig: &mgmtv1alpha1.TransformE164PhoneNumber{
					PreserveLength: &preserveLength,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalNumber := "+12345678901"
		result, err := executor.Mutate(originalNumber, executor.Opts)
		require.NoError(t, err)
		require.NotEqual(t, originalNumber, result)
		require.Equal(t, len(originalNumber), len(*result.(*string)))
	})

	t.Run("TransformE164PhoneNumberConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig{
				TransformE164PhoneNumberConfig: &mgmtv1alpha1.TransformE164PhoneNumber{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalNumber := "+12345678901"
		result, err := executor.Mutate(originalNumber, executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
		require.NotEqual(t, originalNumber, result)
	})

	t.Run("TransformE164PhoneNumberConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalNumber := "+12345678901"
		result, err := executor.Mutate(originalNumber, executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
		require.NotEqual(t, originalNumber, result)
	})

	t.Run("TransformFirstNameConfig", func(t *testing.T) {
		preserveLength := true
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
				TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{
					PreserveLength: &preserveLength,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalName := "John"
		result, err := executor.Mutate(originalName, executor.Opts)
		require.NoError(t, err)
		require.NotEqual(t, originalName, result)
		require.Equal(t, len(originalName), len(*result.(*string)))
	})

	t.Run("TransformFirstNameConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
				TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalName := "John"
		result, err := executor.Mutate(originalName, executor.Opts)
		require.NoError(t, err)
		require.NotEqual(t, originalName, result)
		require.NotEmpty(t, result)
	})

	t.Run("TransformFirstNameConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalName := "John"
		result, err := executor.Mutate(originalName, executor.Opts)
		require.NoError(t, err)
		require.NotEqual(t, originalName, result)
		require.NotEmpty(t, result)
	})

	t.Run("TransformFloat64Config", func(t *testing.T) {
		randomizationRangeMin := float64(-10.0)
		randomizationRangeMax := float64(10.0)
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFloat64Config{
				TransformFloat64Config: &mgmtv1alpha1.TransformFloat64{
					RandomizationRangeMin: &randomizationRangeMin,
					RandomizationRangeMax: &randomizationRangeMax,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalValue := float64(5.5)
		result, err := executor.Mutate(originalValue, executor.Opts)
		require.NoError(t, err)
		transformedValue := *result.(*float64)
		require.NotEqual(t, originalValue, transformedValue)
		require.GreaterOrEqual(t, transformedValue, originalValue+randomizationRangeMin)
		require.LessOrEqual(t, transformedValue, originalValue+randomizationRangeMax)
	})

	t.Run("TransformFloat64Config_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFloat64Config{
				TransformFloat64Config: &mgmtv1alpha1.TransformFloat64{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalValue := float64(5.5)
		result, err := executor.Mutate(originalValue, executor.Opts)
		require.NoError(t, err)
		transformedValue := *result.(*float64)
		require.NotEqual(t, originalValue, transformedValue)
	})

	t.Run("TransformFloat64Config_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFloat64Config{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalValue := float64(5.5)
		result, err := executor.Mutate(originalValue, executor.Opts)
		require.NoError(t, err)
		transformedValue := *result.(*float64)
		require.NotEqual(t, originalValue, transformedValue)
	})

	t.Run("TransformInt64PhoneNumberConfig", func(t *testing.T) {
		preserveLength := true
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig{
				TransformInt64PhoneNumberConfig: &mgmtv1alpha1.TransformInt64PhoneNumber{
					PreserveLength: &preserveLength,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalNumber := int64(1234567890)
		result, err := executor.Mutate(originalNumber, executor.Opts)
		require.NoError(t, err)
		transformedNumber := *result.(*int64)
		require.NotEqual(t, originalNumber, transformedNumber)
		require.GreaterOrEqual(t, transformedNumber, int64(1000000000))
		require.Less(t, transformedNumber, int64(10000000000))
	})

	t.Run("TransformInt64PhoneNumberConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig{
				TransformInt64PhoneNumberConfig: &mgmtv1alpha1.TransformInt64PhoneNumber{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalNumber := int64(1234567890)
		result, err := executor.Mutate(originalNumber, executor.Opts)
		require.NoError(t, err)
		transformedNumber := *result.(*int64)
		require.NotEqual(t, originalNumber, transformedNumber)
		require.NotEmpty(t, transformedNumber)
	})

	t.Run("TransformInt64PhoneNumberConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalNumber := int64(1234567890)
		result, err := executor.Mutate(originalNumber, executor.Opts)
		require.NoError(t, err)
		transformedNumber := *result.(*int64)
		require.NotEqual(t, originalNumber, transformedNumber)
		require.NotEmpty(t, transformedNumber)
	})

	t.Run("TransformLastNameConfig", func(t *testing.T) {
		preserveLength := true
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformLastNameConfig{
				TransformLastNameConfig: &mgmtv1alpha1.TransformLastName{
					PreserveLength: &preserveLength,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalName := "Smith"
		result, err := executor.Mutate(originalName, executor.Opts)
		require.NoError(t, err)
		require.NotEqual(t, originalName, result)
		require.Equal(t, len(originalName), len(*result.(*string)))
	})

	t.Run("TransformLastNameConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformLastNameConfig{
				TransformLastNameConfig: &mgmtv1alpha1.TransformLastName{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalName := "Smith"
		result, err := executor.Mutate(originalName, executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
		require.NotEqual(t, originalName, result)
	})

	t.Run("TransformLastNameConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformLastNameConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalName := "Smith"
		result, err := executor.Mutate(originalName, executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
		require.NotEqual(t, originalName, result)
	})

	t.Run("TransformPhoneNumberConfig", func(t *testing.T) {
		preserveLength := true
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig{
				TransformPhoneNumberConfig: &mgmtv1alpha1.TransformPhoneNumber{
					PreserveLength: &preserveLength,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalNumber := "123-456-7890"
		result, err := executor.Mutate(originalNumber, executor.Opts)
		require.NoError(t, err)
		require.NotEqual(t, originalNumber, result)
		require.Equal(t, len(originalNumber), len(*result.(*string)))
	})

	t.Run("TransformPhoneNumberConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig{
				TransformPhoneNumberConfig: &mgmtv1alpha1.TransformPhoneNumber{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalNumber := "123-456-7890"
		result, err := executor.Mutate(originalNumber, executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
		require.NotEqual(t, originalNumber, result)
	})

	t.Run("TransformPhoneNumberConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalNumber := "123-456-7890"
		result, err := executor.Mutate(originalNumber, executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
		require.NotEqual(t, originalNumber, result)
	})

	t.Run("Nullconfig", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("any value", executor.Opts)
		require.NoError(t, err)
		require.Equal(t, "null", result)
	})

	t.Run("GenerateDefaultConfig", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{
				GenerateDefaultConfig: &mgmtv1alpha1.GenerateDefault{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate("any value", executor.Opts)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})

	t.Run("TransformCharacterScrambleConfig", func(t *testing.T) {
		userProvidedRegex := `[a-zA-Z]`
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
				TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{
					UserProvidedRegex: &userProvidedRegex,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalString := "Hello123World"
		result, err := executor.Mutate(originalString, executor.Opts)
		require.NoError(t, err)
		require.NotEqual(t, originalString, result)
		require.Equal(t, len(originalString), len(*result.(*string)))
	})

	t.Run("TransformCharacterScrambleConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
				TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalString := "Hello123World"
		result, err := executor.Mutate(originalString, executor.Opts)
		require.NoError(t, err)
		require.NotEqual(t, originalString, result)
		require.NotEmpty(t, result)
	})

	t.Run("TransformCharacterScrambleConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		originalString := "Hello123World"
		result, err := executor.Mutate(originalString, executor.Opts)
		require.NoError(t, err)
		require.NotEqual(t, originalString, result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateCountryConfig", func(t *testing.T) {
		genFull := true
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCountryConfig{
				GenerateCountryConfig: &mgmtv1alpha1.GenerateCountry{
					GenerateFullName: &genFull,
				},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
		require.Greater(t, len(result.(string)), 2)
	})

	t.Run("GenerateCountryConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCountryConfig{
				GenerateCountryConfig: &mgmtv1alpha1.GenerateCountry{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
		require.LessOrEqual(t, len(result.(string)), 2)
	})

	t.Run("GenerateCountryConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCountryConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
		require.LessOrEqual(t, len(result.(string)), 2)
	})

	t.Run("TransformPiiTextConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{
				TransformPiiTextConfig: &mgmtv1alpha1.TransformPiiText{},
			},
		}

		mockanalyze := presidioapi.NewMockAnalyzeInterface(t)
		mockanon := presidioapi.NewMockAnonymizeInterface(t)

		mockanalyze.On("PostAnalyzeWithResponse", mock.Anything, mock.Anything).
			Return(&presidioapi.PostAnalyzeResponse{
				JSON200: &[]presidioapi.RecognizerResultWithAnaysisExplanation{
					{},
				},
			}, nil)

		mockText := "bar"
		mockanon.On("PostAnonymizeWithResponse", mock.Anything, mock.Anything).
			Return(&presidioapi.PostAnonymizeResponse{
				JSON200: &presidioapi.AnonymizeResponse{Text: &mockText},
			}, nil)

		execOpts := []TransformerExecutorOption{
			WithTransformPiiTextConfig(mockanalyze, mockanon),
		}
		executor, err := InitializeTransformerByConfigType(config, execOpts...)
		require.NoError(t, err)
		require.NotNil(t, executor)

		originalText := "Hello, John Doe!"
		result, err := executor.Mutate(originalText, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEqual(t, originalText, result)
		require.Equal(t, mockText, result)
	})

	t.Run("TransformPiiTextConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{},
		}

		mockanalyze := presidioapi.NewMockAnalyzeInterface(t)
		mockanon := presidioapi.NewMockAnonymizeInterface(t)

		mockanalyze.On("PostAnalyzeWithResponse", mock.Anything, mock.Anything).
			Return(&presidioapi.PostAnalyzeResponse{
				JSON200: &[]presidioapi.RecognizerResultWithAnaysisExplanation{
					{},
				},
			}, nil)

		mockText := "bar"
		mockanon.On("PostAnonymizeWithResponse", mock.Anything, mock.Anything).
			Return(&presidioapi.PostAnonymizeResponse{
				JSON200: &presidioapi.AnonymizeResponse{Text: &mockText},
			}, nil)

		execOpts := []TransformerExecutorOption{
			WithTransformPiiTextConfig(mockanalyze, mockanon),
		}
		executor, err := InitializeTransformerByConfigType(config, execOpts...)
		require.NoError(t, err)
		require.NotNil(t, executor)

		originalText := "Hello, John Doe!"
		result, err := executor.Mutate(originalText, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEqual(t, originalText, result)
		require.Equal(t, mockText, result)
	})

	t.Run("GenerateBusinessNameConfig_Empty", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateBusinessNameConfig{
				GenerateBusinessNameConfig: &mgmtv1alpha1.GenerateBusinessName{},
			},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("GenerateBusinessNameConfig_Nil", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateBusinessNameConfig{},
		}
		executor, err := InitializeTransformerByConfigType(config)
		require.NoError(t, err)
		require.NotNil(t, executor)
		result, err := executor.Mutate(nil, executor.Opts)
		require.NoError(t, err)
		require.IsType(t, "", result)
		require.NotEmpty(t, result)
	})

	t.Run("UnsupportedConfig", func(t *testing.T) {
		config := &mgmtv1alpha1.TransformerConfig{
			Config: nil,
		}
		_, err := InitializeTransformerByConfigType(config)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported transformer")
	})
}
