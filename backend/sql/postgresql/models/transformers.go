package pg_models

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type Transformer struct {
	Value  string              `json:"value"`
	Config *TransformerConfigs `json:"config,omitempty"`
}

type TransformerConfigs struct {
	GenerateEmail          *GenerateEmailConfig          `json:"generateEmailConfig,omitempty"`
	GenerateRealisticEmail *GenerateRealisticEmailConfig `json:"generateRealisticEmailConfig,omitempty"`
	TransformEmail         *TransformEmailConfig         `json:"transformEmail,omitempty"`
	GenerateBool           *GenerateBoolConfig           `json:"generateBool,omitempty"`
	GenerateCardNumber     *GenerateCardNumberConfig     `json:"generateCardNumber,omitempty"`
	GenerateCity           *GenerateCityConfig           `json:"generateCity,omitempty"`
	GenerateE164Number     *GenerateE164NumberConfig     `json:"generateE164Number,omitempty"`
	GenerateFirstName      *GenerateFirstNameConfig      `json:"generateFirstName,omitempty"`
	GenerateFloat          *GenerateFloatConfig          `json:"generateFloat,omitempty"`
	GenerateFullAddress    *GenerateFullAddressConfig    `json:"generateFullAddress,omitempty"`
	GenerateFullName       *GenerateFullNameConfig       `json:"generateFullName,omitempty"`
	GenerateGender         *GenerateGenderConfig         `json:"generateGender,omitempty"`
	GenerateInt64Phone     *GenerateInt64PhoneConfig     `json:"generateInt64Phone,omitempty"`
	GenerateInt            *GenerateIntConfig            `json:"GenerateInt,omitempty"`
	GenerateLastName       *GenerateLastNameConfig       `json:"generateLastName,omitempty"`
	GenerateSha256Hash     *GenerateSha256HashConfig     `json:"generateSha256Hash,omitempty"`
	GenerateSsn            *GenerateSsnConfig            `json:"generateSsnConfig,omitempty"`
	GenerateState          *GenerateStateConfig          `json:"generateStateConfig,omitempty"`
	GenerateStreetAddress  *GenerateStreetAddressConfig  `json:"generateStreetAddressConfig,omitempty"`
	GenerateStringPhone    *GenerateStringPhoneConfig    `json:"generateStringPhone,omitempty"`
	GenerateString         *GenerateStringConfig         `json:"generateString,omitempty"`
	GenerateUnixTimestamp  *GenerateUnixTimestampConfig  `json:"generateUnixTimestamp,omitempty"`
	GenerateUsername       *GenerateUsernameConfig       `json:"generateUsername,omitempty"`
	GenerateUtcTimestamp   *GenerateUtcTimestampConfig   `json:"generateUtcTimestamp,omitempty"`
	GenerateUuid           *GenerateUuidConfig           `json:"generateUuid,omitempty"`
	GenerateZipcode        *GenerateZipcodeConfig        `json:"generateZipcode,omitempty"`
	TransformE164Phone     *TransformE164PhoneConfig     `json:"transformE164Phone,omitempty"`
	TransformFirstname     *TransformFirstNameConfig     `json:"transformFirstName,omitempty"`
	TransformFloat         *TransformFloatConfig         `json:"transformFloat,omitempty"`
	TransformFullName      *TransformFullNameConfig      `json:"transformFullName,,omitempty"`
	TransformIntPhone      *TransformIntPhoneConfig      `json:"transformIntPhone,omitempty"`
	TransformInt           *TransformIntConfig           `json:"transformInt,omitempty"`
	TransformLastName      *TransformLastNameConfig      `json:"transformLastName,omitempty"`
	TransformPhone         *TransformPhoneConfig         `json:"transformPhone,omitempty"`
	TransformString        *TransformStringConfig        `json:"transformString,omitempty"`
	Passthrough            *PassthroughConfig            `json:"passthrough,omitempty"`
	Null                   *NullConfig                   `json:"null,omitempty"`
}

type GenerateEmailConfig struct{}

type GenerateRealisticEmailConfig struct{}

type TransformEmailConfig struct {
	PreserveLength bool `json:"preserveLength"`
	PreserveDomain bool `json:"preserveDomain"`
}

type GenerateBoolConfig struct{}

type GenerateCardNumberConfig struct {
	ValidLuhn bool `json:"validLuhn"`
}

type GenerateCityConfig struct{}

type GenerateE164NumberConfig struct {
	Length int64 `json:"length"`
}

type GenerateFirstNameConfig struct{}

type GenerateFloatConfig struct {
	Sign                string `json:"sign"`
	DigitsBeforeDecimal int64  `json:"digitsBeforeDecimal"`
	DigitsAfterDecimal  int64  `json:"digitsAfterDecimal"`
}

type GenerateFullAddressConfig struct{}

type GenerateFullNameConfig struct{}

type GenerateGenderConfig struct {
	Abbreviate bool `json:"abbreviate"`
}

type GenerateInt64PhoneConfig struct{}

type GenerateIntConfig struct {
	Length int64  `json:"length"`
	Sign   string `json:"sign"`
}

type GenerateLastNameConfig struct{}

type GenerateSha256HashConfig struct{}

type GenerateSsnConfig struct{}

type GenerateStateConfig struct{}

type GenerateStreetAddressConfig struct{}

type GenerateStringPhoneConfig struct {
	IncludeHyphens bool `json:"includeHyphens"`
}

type GenerateStringConfig struct {
	Length int64 `json:"length"`
}

type GenerateUnixTimestampConfig struct{}

type GenerateUsernameConfig struct{}

type GenerateUtcTimestampConfig struct{}

type GenerateUuidConfig struct {
	IncludeHyphens bool `json:"includeHyphens"`
}

type GenerateZipcodeConfig struct{}

type TransformE164PhoneConfig struct {
	PreserveLength bool `json:"preserveLength"`
}

type TransformFirstNameConfig struct {
	PreserveLength bool `json:"preserveLength"`
}

type TransformFloatConfig struct {
	PreserveLength bool `json:"preserveLength"`
	PreserveSign   bool `json:"preserveSign"`
}

type TransformFullNameConfig struct {
	PreserveLength bool `json:"preserveLength"`
}

type TransformIntPhoneConfig struct {
	PreserveLength bool `json:"preserveLength"`
}

type TransformIntConfig struct {
	PreserveLength bool `json:"preserveLength"`
	PreserveSign   bool `json:"preserveSign"`
}

type TransformLastNameConfig struct {
	PreserveLength bool `json:"preserveLength"`
}

type TransformPhoneConfig struct {
	PreserveLength bool `json:"preserveLength"`
	IncludeHyphens bool `json:"includeHyphens"`
}

type TransformStringConfig struct {
	PreserveLength bool `json:"preserveLength"`
}

type PassthroughConfig struct{}

type NullConfig struct{}

// from API -> DB
func (t *Transformer) FromTransformerDto(tr *mgmtv1alpha1.Transformer) error {

	t.Value = tr.Value

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
	case *mgmtv1alpha1.TransformerConfig_GenerateRealisticEmailConfig:
		t.GenerateRealisticEmail = &GenerateRealisticEmailConfig{}
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
	case *mgmtv1alpha1.TransformerConfig_GenerateE164NumberConfig:
		t.GenerateE164Number = &GenerateE164NumberConfig{
			Length: tr.GetGenerateE164NumberConfig().Length,
		}
	case *mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig:
		t.GenerateFirstName = &GenerateFirstNameConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateFloatConfig:
		t.GenerateFloat = &GenerateFloatConfig{
			Sign:                tr.GetGenerateFloatConfig().Sign,
			DigitsBeforeDecimal: tr.GetGenerateFloatConfig().DigitsBeforeDecimal,
			DigitsAfterDecimal:  tr.GetGenerateFloatConfig().DigitsAfterDecimal,
		}
	case *mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig:
		t.GenerateFullAddress = &GenerateFullAddressConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig:
		t.GenerateFullName = &GenerateFullNameConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateGenderConfig:
		t.GenerateGender = &GenerateGenderConfig{
			Abbreviate: tr.GetGenerateGenderConfig().Abbreviate,
		}
	case *mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneConfig:
		t.GenerateInt64Phone = &GenerateInt64PhoneConfig{}
	case *mgmtv1alpha1.TransformerConfig_GenerateIntConfig:
		t.GenerateInt = &GenerateIntConfig{
			Length: tr.GetGenerateIntConfig().Length,
			Sign:   tr.GetGenerateIntConfig().Sign,
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
	case *mgmtv1alpha1.TransformerConfig_GenerateStringPhoneConfig:
		t.GenerateStringPhone = &GenerateStringPhoneConfig{
			IncludeHyphens: tr.GetGenerateStringPhoneConfig().IncludeHyphens,
		}
	case *mgmtv1alpha1.TransformerConfig_GenerateStringConfig:
		t.GenerateString = &GenerateStringConfig{
			Length: tr.GetGenerateStringConfig().Length,
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
	case *mgmtv1alpha1.TransformerConfig_TransformE164PhoneConfig:
		t.TransformE164Phone = &TransformE164PhoneConfig{
			PreserveLength: tr.GetTransformE164PhoneConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig:
		t.TransformFirstname = &TransformFirstNameConfig{
			PreserveLength: tr.GetTransformFirstNameConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformFloatConfig:
		t.TransformFloat = &TransformFloatConfig{
			PreserveLength: tr.GetTransformFloatConfig().PreserveLength,
			PreserveSign:   tr.GetTransformFloatConfig().PreserveSign,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformFullNameConfig:
		t.TransformFullName = &TransformFullNameConfig{
			PreserveLength: tr.GetTransformFullNameConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformIntPhoneConfig:
		t.TransformIntPhone = &TransformIntPhoneConfig{
			PreserveLength: tr.GetTransformIntConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformIntConfig:
		t.TransformInt = &TransformIntConfig{
			PreserveLength: tr.GetTransformIntConfig().PreserveLength,
			PreserveSign:   tr.GetTransformIntConfig().PreserveSign,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformLastNameConfig:
		t.TransformLastName = &TransformLastNameConfig{
			PreserveLength: tr.GetTransformLastNameConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformPhoneConfig:
		t.TransformPhone = &TransformPhoneConfig{
			PreserveLength: tr.GetTransformPhoneConfig().PreserveLength,
			IncludeHyphens: tr.GetTransformPhoneConfig().IncludeHyphens,
		}
	case *mgmtv1alpha1.TransformerConfig_TransformStringConfig:
		t.TransformString = &TransformStringConfig{
			PreserveLength: tr.GetTransformStringConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_PassthroughConfig:
		t.Passthrough = &PassthroughConfig{}
	case *mgmtv1alpha1.TransformerConfig_Nullconfig:
		t.Null = &NullConfig{}
	default:
		t = &TransformerConfigs{}
	}

	return nil
}

// DB -> API

func (t *Transformer) ToTransformerDto() *mgmtv1alpha1.Transformer {

	config := &TransformerConfigs{}

	return &mgmtv1alpha1.Transformer{
		Value:  t.Value,
		Config: config.ToTransformerConfigDto(t.Config),
	}
}

func (t *TransformerConfigs) ToTransformerConfigDto(tr *TransformerConfigs) *mgmtv1alpha1.TransformerConfig {
	switch {
	case tr.GenerateEmail != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
				GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{},
			},
		}
	case tr.GenerateRealisticEmail != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateRealisticEmailConfig{
				GenerateRealisticEmailConfig: &mgmtv1alpha1.GenerateRealisticEmail{},
			},
		}
	case tr.TransformEmail != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
				TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
					PreserveDomain: tr.TransformEmail.PreserveDomain,
					PreserveLength: tr.TransformEmail.PreserveLength,
				},
			},
		}
	case tr.GenerateBool != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
				GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
			},
		}
	case tr.GenerateCardNumber != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{
				GenerateCardNumberConfig: &mgmtv1alpha1.GenerateCardNumber{
					ValidLuhn: tr.GenerateCardNumber.ValidLuhn,
				},
			},
		}
	case tr.GenerateCity != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
				GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
			},
		}
	case tr.GenerateE164Number != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateE164NumberConfig{
				GenerateE164NumberConfig: &mgmtv1alpha1.GenerateE164Number{
					Length: tr.GenerateE164Number.Length,
				},
			},
		}
	case tr.GenerateFirstName != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
				GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
			},
		}
	case tr.GenerateFloat != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFloatConfig{
				GenerateFloatConfig: &mgmtv1alpha1.GenerateFloat{
					Sign:                tr.GenerateFloat.Sign,
					DigitsBeforeDecimal: tr.GenerateFloat.DigitsBeforeDecimal,
					DigitsAfterDecimal:  tr.GenerateFloat.DigitsAfterDecimal,
				},
			},
		}
	case tr.GenerateFullAddress != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig{
				GenerateFullAddressConfig: &mgmtv1alpha1.GenerateFullAddress{},
			},
		}
	case tr.GenerateFullName != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
				GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
			},
		}
	case tr.GenerateGender != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateGenderConfig{
				GenerateGenderConfig: &mgmtv1alpha1.GenerateGender{
					Abbreviate: tr.GenerateGender.Abbreviate,
				},
			},
		}
	case tr.GenerateInt64Phone != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneConfig{
				GenerateInt64PhoneConfig: &mgmtv1alpha1.GenerateInt64Phone{},
			},
		}
	case tr.GenerateInt != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateIntConfig{
				GenerateIntConfig: &mgmtv1alpha1.GenerateInt{
					Length: tr.GenerateInt.Length,
					Sign:   tr.GenerateInt.Sign,
				},
			},
		}
	case tr.GenerateLastName != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig{
				GenerateLastNameConfig: &mgmtv1alpha1.GenerateLastName{},
			},
		}
	case tr.GenerateSha256Hash != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig{
				GenerateSha256HashConfig: &mgmtv1alpha1.GenerateSha256Hash{},
			},
		}
	case tr.GenerateSsn != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
				GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
			},
		}
	case tr.GenerateState != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStateConfig{
				GenerateStateConfig: &mgmtv1alpha1.GenerateState{},
			},
		}
	case tr.GenerateStreetAddress != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig{
				GenerateStreetAddressConfig: &mgmtv1alpha1.GenerateStreetAddress{},
			},
		}
	case tr.GenerateStringPhone != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStringPhoneConfig{
				GenerateStringPhoneConfig: &mgmtv1alpha1.GenerateStringPhone{
					IncludeHyphens: tr.GenerateStringPhone.IncludeHyphens,
				},
			},
		}
	case tr.GenerateString != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
				GenerateStringConfig: &mgmtv1alpha1.GenerateString{
					Length: tr.GenerateString.Length,
				},
			},
		}
	case tr.GenerateUnixTimestamp != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{
				GenerateUnixtimestampConfig: &mgmtv1alpha1.GenerateUnixTimestamp{},
			},
		}
	case tr.GenerateUsername != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig{
				GenerateUsernameConfig: &mgmtv1alpha1.GenerateUsername{},
			},
		}
	case tr.GenerateUtcTimestamp != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig{
				GenerateUtctimestampConfig: &mgmtv1alpha1.GenerateUtcTimestamp{},
			},
		}
	case tr.GenerateUuid != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
				GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
					IncludeHyphens: tr.GenerateUuid.IncludeHyphens,
				},
			},
		}
	case tr.GenerateZipcode != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig{
				GenerateZipcodeConfig: &mgmtv1alpha1.GenerateZipcode{},
			},
		}
	case tr.TransformE164Phone != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformE164PhoneConfig{
				TransformE164PhoneConfig: &mgmtv1alpha1.TransformE164Phone{
					PreserveLength: tr.TransformE164Phone.PreserveLength,
				},
			},
		}
	case tr.TransformFirstname != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
				TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{
					PreserveLength: tr.TransformFirstname.PreserveLength,
				},
			},
		}
	case tr.TransformFloat != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFloatConfig{
				TransformFloatConfig: &mgmtv1alpha1.TransformFloat{
					PreserveLength: tr.TransformFloat.PreserveLength,
					PreserveSign:   tr.TransformFloat.PreserveSign,
				},
			},
		}
	case tr.TransformFullName != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
				TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{
					PreserveLength: tr.TransformFullName.PreserveLength,
				},
			},
		}
	case tr.TransformIntPhone != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformIntPhoneConfig{
				TransformIntPhoneConfig: &mgmtv1alpha1.TransformIntPhone{
					PreserveLength: tr.TransformInt.PreserveLength,
				},
			},
		}
	case tr.TransformInt != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformIntConfig{
				TransformIntConfig: &mgmtv1alpha1.TransformInt{
					PreserveLength: tr.TransformInt.PreserveLength,
					PreserveSign:   tr.TransformInt.PreserveSign,
				},
			},
		}
	case tr.TransformLastName != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformLastNameConfig{
				TransformLastNameConfig: &mgmtv1alpha1.TransformLastName{
					PreserveLength: tr.TransformLastName.PreserveLength,
				},
			},
		}
	case tr.TransformPhone != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformPhoneConfig{
				TransformPhoneConfig: &mgmtv1alpha1.TransformPhone{
					PreserveLength: tr.TransformPhone.PreserveLength,
					IncludeHyphens: tr.TransformPhone.IncludeHyphens,
				},
			},
		}
	case tr.TransformString != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
				TransformStringConfig: &mgmtv1alpha1.TransformString{
					PreserveLength: tr.TransformString.PreserveLength,
				},
			},
		}
	case tr.Passthrough != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
				PassthroughConfig: &mgmtv1alpha1.Passthrough{},
			},
		}
	case tr.Null != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		}
	default:
		return &mgmtv1alpha1.TransformerConfig{}
	}
}
