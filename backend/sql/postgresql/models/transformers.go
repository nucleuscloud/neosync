package pg_models

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type JobMappingTransformerModel struct {
	Source string              `json:"source"`
	Config *TransformerConfigs `json:"config,omitempty"`
}

type TransformerConfigs struct {
	GenerateEmail             *GenerateEmailConfig             `json:"generateEmailConfig,omitempty"`
	TransformEmail            *TransformEmailConfig            `json:"transformEmail,omitempty"`
	GenerateBool              *GenerateBoolConfig              `json:"generateBool,omitempty"`
	GenerateCardNumber        *GenerateCardNumberConfig        `json:"generateCardNumber,omitempty"`
	GenerateCity              *GenerateCityConfig              `json:"generateCity,omitempty"`
	GenerateDefault           *GenerateDefaultConfig           `json:"generateDefault,omitempty"`
	GenerateE164PhoneNumber   *GenerateE164PhoneNumberConfig   `json:"generateE164PhoneNumber,omitempty"`
	GenerateFirstName         *GenerateFirstNameConfig         `json:"generateFirstName,omitempty"`
	GenerateFloat64           *GenerateFloat64Config           `json:"generateFloat64,omitempty"`
	GenerateFullAddress       *GenerateFullAddressConfig       `json:"generateFullAddress,omitempty"`
	GenerateFullName          *GenerateFullNameConfig          `json:"generateFullName,omitempty"`
	GenerateGender            *GenerateGenderConfig            `json:"generateGender,omitempty"`
	GenerateInt64PhoneNumber  *GenerateInt64PhoneNumberConfig  `json:"generateInt64PhoneNumber,omitempty"`
	GenerateInt64             *GenerateInt64Config             `json:"GenerateInt64,omitempty"`
	GenerateLastName          *GenerateLastNameConfig          `json:"generateLastName,omitempty"`
	GenerateSha256Hash        *GenerateSha256HashConfig        `json:"generateSha256Hash,omitempty"`
	GenerateSsn               *GenerateSsnConfig               `json:"generateSsnConfig,omitempty"`
	GenerateState             *GenerateStateConfig             `json:"generateStateConfig,omitempty"`
	GenerateStreetAddress     *GenerateStreetAddressConfig     `json:"generateStreetAddressConfig,omitempty"`
	GenerateStringPhoneNumber *GenerateStringPhoneNumberConfig `json:"generateStringPhoneNumber,omitempty"`
	GenerateString            *GenerateStringConfig            `json:"generateString,omitempty"`
	GenerateUnixTimestamp     *GenerateUnixTimestampConfig     `json:"generateUnixTimestamp,omitempty"`
	GenerateUsername          *GenerateUsernameConfig          `json:"generateUsername,omitempty"`
	GenerateUtcTimestamp      *GenerateUtcTimestampConfig      `json:"generateUtcTimestamp,omitempty"`
	GenerateUuid              *GenerateUuidConfig              `json:"generateUuid,omitempty"`
	GenerateZipcode           *GenerateZipcodeConfig           `json:"generateZipcode,omitempty"`
	TransformE164PhoneNumber  *TransformE164PhoneNumberConfig  `json:"transformE164PhoneNumber,omitempty"`
	TransformFirstname        *TransformFirstNameConfig        `json:"transformFirstName,omitempty"`
	TransformFloat64          *TransformFloat64Config          `json:"transformFloat64,omitempty"`
	TransformFullName         *TransformFullNameConfig         `json:"transformFullName,,omitempty"`
	TransformInt64PhoneNumber *TransformInt64PhoneNumberConfig `json:"transformInt64PhoneNumber,omitempty"`
	TransformInt64            *TransformInt64Config            `json:"transformInt64,omitempty"`
	TransformLastName         *TransformLastNameConfig         `json:"transformLastName,omitempty"`
	TransformPhoneNumber      *TransformPhoneNumberConfig      `json:"transformPhoneNumber,omitempty"`
	TransformString           *TransformStringConfig           `json:"transformString,omitempty"`
	Passthrough               *PassthroughConfig               `json:"passthrough,omitempty"`
	Null                      *NullConfig                      `json:"null,omitempty"`
	UserDefinedTransformer    *UserDefinedTransformerConfig    `json:"userDefinedTransformer,omitempty"`
	TransformJavascript       *TransformJavascriptConfig       `json:"transformJavascript,omitempty"`
}

type GenerateEmailConfig struct{}

type TransformEmailConfig struct {
	PreserveLength bool `json:"preserveLength"`
	PreserveDomain bool `json:"preserveDomain"`
}

type GenerateBoolConfig struct{}

type GenerateCardNumberConfig struct {
	ValidLuhn bool `json:"validLuhn"`
}

type GenerateCityConfig struct{}

type GenerateDefaultConfig struct{}

type GenerateE164PhoneNumberConfig struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

type GenerateFirstNameConfig struct{}

type GenerateFloat64Config struct {
	RandomizeSign bool    `json:"randomizeSign"`
	Min           float64 `json:"min"`
	Max           float64 `json:"max"`
	Precision     int64   `json:"precision"`
}

type GenerateFullAddressConfig struct{}

type GenerateFullNameConfig struct{}

type GenerateGenderConfig struct {
	Abbreviate bool `json:"abbreviate"`
}

type GenerateInt64PhoneNumberConfig struct{}

type GenerateInt64Config struct {
	RandomizeSign bool  `json:"randomizeSign"`
	Min           int64 `json:"min"`
	Max           int64 `json:"max"`
}

type GenerateLastNameConfig struct{}

type GenerateSha256HashConfig struct{}

type GenerateSsnConfig struct{}

type GenerateStateConfig struct{}

type GenerateStreetAddressConfig struct{}

type GenerateStringPhoneNumberConfig struct {
	IncludeHyphens bool `json:"includeHyphens"`
}

type GenerateStringConfig struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

type GenerateUnixTimestampConfig struct{}

type GenerateUsernameConfig struct{}

type GenerateUtcTimestampConfig struct{}

type GenerateUuidConfig struct {
	IncludeHyphens bool `json:"includeHyphens"`
}

type GenerateZipcodeConfig struct{}

type TransformE164PhoneNumberConfig struct {
	PreserveLength bool `json:"preserveLength"`
}

type TransformFirstNameConfig struct {
	PreserveLength bool `json:"preserveLength"`
}

type TransformFloat64Config struct {
	RandomizationRangeMin float64 `json:"randomizationRangeMin"`
	RandomizationRangeMax float64 `json:"randomizationRangeMax"`
}

type TransformFullNameConfig struct {
	PreserveLength bool `json:"preserveLength"`
}

type TransformInt64PhoneNumberConfig struct {
	PreserveLength bool `json:"preserveLength"`
}

type TransformInt64Config struct {
	RandomizationRangeMin int64 `json:"randomizationRangeMin"`
	RandomizationRangeMax int64 `json:"randomizationRangeMax"`
}

type TransformLastNameConfig struct {
	PreserveLength bool `json:"preserveLength"`
}

type TransformPhoneNumberConfig struct {
	PreserveLength bool `json:"preserveLength"`
	IncludeHyphens bool `json:"includeHyphens"`
}

type TransformStringConfig struct {
	PreserveLength bool `json:"preserveLength"`
}

type PassthroughConfig struct{}

type NullConfig struct{}

type UserDefinedTransformerConfig struct {
	Id string `json:"id"`
}

type TransformJavascriptConfig struct {
	Code string `json:"code"`
}

// from API -> DB
func (t *JobMappingTransformerModel) FromTransformerDto(tr *mgmtv1alpha1.JobMappingTransformer) error {

	t.Source = tr.Source

	config := &TransformerConfigs{}

	if err := config.FromTransformerConfigDto(tr.Config); err != nil {
		return err
	}

	t.Config = config

	return nil
}

func (t *TransformerConfigs) FromTransformerConfigDto(tr *mgmtv1alpha1.TransformerConfig) error {

	switch tr.Config.(type) {
	case *mgmtv1alpha1.TransformerConfig_GenerateEmailConfig:
		t.GenerateEmail = &GenerateEmailConfig{}
	case *mgmtv1alpha1.TransformerConfig_TransformEmailConfig:
		t.TransformEmail = &TransformEmailConfig{
			PreserveLength: tr.GetTransformEmailConfig().PreserveLength,
			PreserveDomain: tr.GetTransformEmailConfig().PreserveDomain,
		}
	case *mgmtv1alpha1.TransformerConfig_GenerateBoolConfig:
		t.GenerateBool = &GenerateBoolConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig:
		t.GenerateCardNumber = &GenerateCardNumberConfig{
			ValidLuhn: tr.GetGenerateCardNumberConfig().ValidLuhn,
		}
	case *mgmtv1alpha1.TransformerConfig_GenerateCityConfig:
		t.GenerateCity = &GenerateCityConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig:
		t.GenerateDefault = &GenerateDefaultConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateE164PhoneNumberConfig:
		t.GenerateE164PhoneNumber = &GenerateE164PhoneNumberConfig{
			Min: tr.GetGenerateE164PhoneNumberConfig().Min,
			Max: tr.GetGenerateE164PhoneNumberConfig().Max,
		}
	case *mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig:
		t.GenerateFirstName = &GenerateFirstNameConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateFloat64Config:
		t.GenerateFloat64 = &GenerateFloat64Config{
			RandomizeSign: tr.GetGenerateFloat64Config().RandomizeSign,
			Min:           tr.GetGenerateFloat64Config().Min,
			Max:           tr.GetGenerateFloat64Config().Max,
			Precision:     tr.GetGenerateFloat64Config().Precision,
		}
	case *mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig:
		t.GenerateFullAddress = &GenerateFullAddressConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig:
		t.GenerateFullName = &GenerateFullNameConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateGenderConfig:
		t.GenerateGender = &GenerateGenderConfig{
			Abbreviate: tr.GetGenerateGenderConfig().Abbreviate,
		}
	case *mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneNumberConfig:
		t.GenerateInt64PhoneNumber = &GenerateInt64PhoneNumberConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateInt64Config:
		t.GenerateInt64 = &GenerateInt64Config{
			RandomizeSign: tr.GetGenerateInt64Config().RandomizeSign,
			Min:           tr.GetGenerateInt64Config().Min,
			Max:           tr.GetGenerateInt64Config().Max,
		}
	case *mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig:
		t.GenerateLastName = &GenerateLastNameConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig:
		t.GenerateSha256Hash = &GenerateSha256HashConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateSsnConfig:
		t.GenerateSsn = &GenerateSsnConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateStateConfig:
		t.GenerateState = &GenerateStateConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig:
		t.GenerateStreetAddress = &GenerateStreetAddressConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig:
		t.GenerateStringPhoneNumber = &GenerateStringPhoneNumberConfig{
			IncludeHyphens: tr.GetGenerateStringPhoneNumberConfig().IncludeHyphens,
		}
	case *mgmtv1alpha1.TransformerConfig_GenerateStringConfig:
		t.GenerateString = &GenerateStringConfig{
			Min: tr.GetGenerateStringConfig().Min,
			Max: tr.GetGenerateStringConfig().Max,
		}
	case *mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig:
		t.GenerateUnixTimestamp = &GenerateUnixTimestampConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig:
		t.GenerateUsername = &GenerateUsernameConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig:
		t.GenerateUtcTimestamp = &GenerateUtcTimestampConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateUuidConfig:
		t.GenerateUuid = &GenerateUuidConfig{
			IncludeHyphens: tr.GetGenerateUuidConfig().IncludeHyphens,
		}
	case *mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig:
		t.GenerateZipcode = &GenerateZipcodeConfig{}
	case *mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig:
		t.TransformE164PhoneNumber = &TransformE164PhoneNumberConfig{
			PreserveLength: tr.GetTransformE164PhoneNumberConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig:
		t.TransformFirstname = &TransformFirstNameConfig{
			PreserveLength: tr.GetTransformFirstNameConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformFloat64Config:
		t.TransformFloat64 = &TransformFloat64Config{
			RandomizationRangeMin: tr.GetTransformFloat64Config().RandomizationRangeMin,
			RandomizationRangeMax: tr.GetTransformFloat64Config().RandomizationRangeMax,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformFullNameConfig:
		t.TransformFullName = &TransformFullNameConfig{
			PreserveLength: tr.GetTransformFullNameConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig:
		t.TransformInt64PhoneNumber = &TransformInt64PhoneNumberConfig{
			PreserveLength: tr.GetTransformInt64PhoneNumberConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformInt64Config:
		t.TransformInt64 = &TransformInt64Config{
			RandomizationRangeMin: tr.GetTransformInt64Config().RandomizationRangeMin,
			RandomizationRangeMax: tr.GetTransformInt64Config().RandomizationRangeMax,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformLastNameConfig:
		t.TransformLastName = &TransformLastNameConfig{
			PreserveLength: tr.GetTransformLastNameConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig:
		t.TransformPhoneNumber = &TransformPhoneNumberConfig{
			PreserveLength: tr.GetTransformPhoneNumberConfig().PreserveLength,
			IncludeHyphens: tr.GetTransformPhoneNumberConfig().IncludeHyphens,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformStringConfig:
		t.TransformString = &TransformStringConfig{
			PreserveLength: tr.GetTransformStringConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_PassthroughConfig:
		t.Passthrough = &PassthroughConfig{}
	case *mgmtv1alpha1.TransformerConfig_Nullconfig:
		t.Null = &NullConfig{}
	case *mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig:
		t.UserDefinedTransformer = &UserDefinedTransformerConfig{
			Id: tr.GetUserDefinedTransformerConfig().Id,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig:
		t.TransformJavascript = &TransformJavascriptConfig{
			Code: tr.GetTransformJavascriptConfig().Code,
		}
	default:
		t = &TransformerConfigs{}
	}

	return nil
}

// DB -> API

func (t *JobMappingTransformerModel) ToTransformerDto() *mgmtv1alpha1.JobMappingTransformer {
	return &mgmtv1alpha1.JobMappingTransformer{
		Source: t.Source,
		Config: t.Config.ToTransformerConfigDto(),
	}
}

func (t *TransformerConfigs) ToTransformerConfigDto() *mgmtv1alpha1.TransformerConfig {
	switch {
	case t.GenerateEmail != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
				GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{},
			},
		}
	case t.TransformEmail != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
				TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
					PreserveDomain: t.TransformEmail.PreserveDomain,
					PreserveLength: t.TransformEmail.PreserveLength,
				},
			},
		}
	case t.GenerateBool != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
				GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
			},
		}
	case t.GenerateCardNumber != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{
				GenerateCardNumberConfig: &mgmtv1alpha1.GenerateCardNumber{
					ValidLuhn: t.GenerateCardNumber.ValidLuhn,
				},
			},
		}
	case t.GenerateCity != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
				GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
			},
		}
	case t.GenerateDefault != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{},
		}
	case t.GenerateE164PhoneNumber != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateE164PhoneNumberConfig{
				GenerateE164PhoneNumberConfig: &mgmtv1alpha1.GenerateE164PhoneNumber{
					Min: t.GenerateE164PhoneNumber.Min,
					Max: t.GenerateE164PhoneNumber.Max,
				},
			},
		}
	case t.GenerateFirstName != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
				GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
			},
		}
	case t.GenerateFloat64 != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{
				GenerateFloat64Config: &mgmtv1alpha1.GenerateFloat64{
					RandomizeSign: t.GenerateFloat64.RandomizeSign,
					Min:           t.GenerateFloat64.Min,
					Max:           t.GenerateFloat64.Max,
					Precision:     t.GenerateFloat64.Precision,
				},
			},
		}
	case t.GenerateFullAddress != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig{
				GenerateFullAddressConfig: &mgmtv1alpha1.GenerateFullAddress{},
			},
		}
	case t.GenerateFullName != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
				GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
			},
		}
	case t.GenerateGender != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateGenderConfig{
				GenerateGenderConfig: &mgmtv1alpha1.GenerateGender{
					Abbreviate: t.GenerateGender.Abbreviate,
				},
			},
		}
	case t.GenerateInt64PhoneNumber != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneNumberConfig{
				GenerateInt64PhoneNumberConfig: &mgmtv1alpha1.GenerateInt64PhoneNumber{},
			},
		}
	case t.GenerateInt64 != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
				GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
					RandomizeSign: t.GenerateInt64.RandomizeSign,
					Min:           t.GenerateInt64.Min,
					Max:           t.GenerateInt64.Max,
				},
			},
		}
	case t.GenerateLastName != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig{
				GenerateLastNameConfig: &mgmtv1alpha1.GenerateLastName{},
			},
		}
	case t.GenerateSha256Hash != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig{
				GenerateSha256HashConfig: &mgmtv1alpha1.GenerateSha256Hash{},
			},
		}
	case t.GenerateSsn != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
				GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
			},
		}
	case t.GenerateState != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStateConfig{
				GenerateStateConfig: &mgmtv1alpha1.GenerateState{},
			},
		}
	case t.GenerateStreetAddress != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig{
				GenerateStreetAddressConfig: &mgmtv1alpha1.GenerateStreetAddress{},
			},
		}
	case t.GenerateStringPhoneNumber != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig{
				GenerateStringPhoneNumberConfig: &mgmtv1alpha1.GenerateStringPhoneNumber{
					IncludeHyphens: t.GenerateStringPhoneNumber.IncludeHyphens,
				},
			},
		}
	case t.GenerateString != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
				GenerateStringConfig: &mgmtv1alpha1.GenerateString{
					Min: t.GenerateString.Min,
					Max: t.GenerateString.Max,
				},
			},
		}
	case t.GenerateUnixTimestamp != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{
				GenerateUnixtimestampConfig: &mgmtv1alpha1.GenerateUnixTimestamp{},
			},
		}
	case t.GenerateUsername != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig{
				GenerateUsernameConfig: &mgmtv1alpha1.GenerateUsername{},
			},
		}
	case t.GenerateUtcTimestamp != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig{
				GenerateUtctimestampConfig: &mgmtv1alpha1.GenerateUtcTimestamp{},
			},
		}
	case t.GenerateUuid != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
				GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
					IncludeHyphens: t.GenerateUuid.IncludeHyphens,
				},
			},
		}
	case t.GenerateZipcode != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig{
				GenerateZipcodeConfig: &mgmtv1alpha1.GenerateZipcode{},
			},
		}
	case t.TransformE164PhoneNumber != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig{
				TransformE164PhoneNumberConfig: &mgmtv1alpha1.TransformE164PhoneNumber{
					PreserveLength: t.TransformE164PhoneNumber.PreserveLength,
				},
			},
		}
	case t.TransformFirstname != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
				TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{
					PreserveLength: t.TransformFirstname.PreserveLength,
				},
			},
		}
	case t.TransformFloat64 != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFloat64Config{
				TransformFloat64Config: &mgmtv1alpha1.TransformFloat64{
					RandomizationRangeMin: t.TransformFloat64.RandomizationRangeMin,
					RandomizationRangeMax: t.TransformFloat64.RandomizationRangeMin,
				},
			},
		}
	case t.TransformFullName != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
				TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{
					PreserveLength: t.TransformFullName.PreserveLength,
				},
			},
		}
	case t.TransformInt64PhoneNumber != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig{
				TransformInt64PhoneNumberConfig: &mgmtv1alpha1.TransformInt64PhoneNumber{
					PreserveLength: t.TransformInt64PhoneNumber.PreserveLength,
				},
			},
		}
	case t.TransformInt64 != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{
				TransformInt64Config: &mgmtv1alpha1.TransformInt64{
					RandomizationRangeMin: t.TransformInt64.RandomizationRangeMin,
					RandomizationRangeMax: t.TransformInt64.RandomizationRangeMax,
				},
			},
		}
	case t.TransformLastName != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformLastNameConfig{
				TransformLastNameConfig: &mgmtv1alpha1.TransformLastName{
					PreserveLength: t.TransformLastName.PreserveLength,
				},
			},
		}
	case t.TransformPhoneNumber != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig{
				TransformPhoneNumberConfig: &mgmtv1alpha1.TransformPhoneNumber{
					PreserveLength: t.TransformPhoneNumber.PreserveLength,
					IncludeHyphens: t.TransformPhoneNumber.IncludeHyphens,
				},
			},
		}
	case t.TransformString != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
				TransformStringConfig: &mgmtv1alpha1.TransformString{
					PreserveLength: t.TransformString.PreserveLength,
				},
			},
		}
	case t.Passthrough != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
				PassthroughConfig: &mgmtv1alpha1.Passthrough{},
			},
		}
	case t.Null != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		}
	case t.UserDefinedTransformer != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig{
				UserDefinedTransformerConfig: &mgmtv1alpha1.UserDefinedTransformerConfig{
					Id: t.UserDefinedTransformer.Id,
				},
			},
		}
	case t.TransformJavascript != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: t.TransformJavascript.Code,
				},
			},
		}
	default:
		return &mgmtv1alpha1.TransformerConfig{}
	}
}
