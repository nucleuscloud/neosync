package pg_models

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type Transformer struct {
	Value  string              `json:"value"`
	Config *TransformerConfigs `json:"config,omitempty"`
}

type TransformerConfigs struct {
	EmailConfig    *EmailConfigs         `json:"emailConfig,omitempty"`
	FirstName      *FirstNameConfig      `json:"firstName,omitempty"`
	LastName       *LastNameConfig       `json:"lastName,omitempty"`
	FullName       *FullNameConfig       `json:"fullName,omitempty"`
	Uuid           *UuidConfig           `json:"uuid,omitempty"`
	PhoneNumber    *PhoneNumberConfig    `json:"phoneNumber,omitempty"`
	IntPhoneNumber *IntPhoneNumberConfig `json:"intPhoneNumber,omitempty"`
	Passthrough    *PassthroughConfig    `json:"passthrough,omitempty"`
	Null           *NullConfig           `json:"null,omitempty"`
	RandomString   *RandomStringConfig   `json:"randomString,omitempty"`
	RandomBool     *RandomBoolConfig     `json:"randomBool,omitempty"`
	RandomInt      *RandomIntConfig      `json:"randomInt,omitempty"`
	RandomFloat    *RandomFloatConfig    `json:"randomFloat,omitempty"`
	Gender         *GenderConfig         `json:"gender,omitempty"`
	UTCTimestamp   *UTCTimestampConfig   `json:"utcTimestamp,omitempty"`
	UnixTimestamp  *UnixTimestampConfig  `json:"unixTimestamp,omitempty"`
	StreetAddress  *StreetAddressConfig  `json:"streetAddress,omitempty"`
	City           *CityConfig           `json:"city,omitempty"`
	Zipcode        *ZipcodeConfig        `json:"zipcode,omitempty"`
	State          *StateConfig          `json:"state,omitempty"`
	FullAddress    *FullAddressConfig    `json:"fullAddress,omitempty"`
	CreditCard     *CreditCardConfig     `json:"creditcard,omitempty"`
	SHA256Hash     *SHA256HashConfig     `json:"sha256Hash,omitempty"`
}

type EmailConfigs struct {
	PreserveLength bool `json:"preserveLength"`
	PreserveDomain bool `json:"preserveDomain"`
}

type FirstNameConfig struct {
	PreserveLength bool `json:"preserveLength"`
}

type LastNameConfig struct {
	PreserveLength bool `json:"preserveLength"`
}

type FullNameConfig struct {
	PreserveLength bool `json:"preserveLength"`
}
type UuidConfig struct {
	IncludeHyphen bool `json:"includeHyphen"`
}
type PhoneNumberConfig struct {
	IncludeHyphens bool `json:"includeHyphens"`
	E164Format     bool `json:"e164Format"`
	PreserveLength bool `json:"preserveLength"`
}

type IntPhoneNumberConfig struct {
	PreserveLength bool `json:"preserveLength"`
}
type PassthroughConfig struct {
}

type NullConfig struct{}

type RandomStringConfig struct {
	PreserveLength bool  `json:"preserveLength"`
	StrLength      int64 `json:"strLength"`
}

type RandomBoolConfig struct{}

type RandomIntConfig struct {
	PreserveLength bool  `json:"preserveLength"`
	IntLength      int64 `json:"intLength"`
}

type RandomFloatConfig struct {
	PreserveLength      bool  `json:"preserveLength"`
	DigitsBeforeDecimal int64 `json:"digitsBeforeDecimal"`
	DigitsAfterDecimal  int64 `json:"digitsAfterDecimal"`
}

type GenderConfig struct {
	Abbreviate bool `json:"abbreviate"`
}

type UTCTimestampConfig struct{}

type UnixTimestampConfig struct{}

type StreetAddressConfig struct{}

type CityConfig struct{}

type ZipcodeConfig struct{}

type StateConfig struct{}

type FullAddressConfig struct{}

type CreditCardConfig struct {
	ValidLuhn bool `json:"validLuhn"`
}

type SHA256HashConfig struct{}

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
	case *mgmtv1alpha1.TransformerConfig_EmailConfig:
		t.EmailConfig = &EmailConfigs{
			PreserveLength: tr.GetEmailConfig().PreserveLength,
			PreserveDomain: tr.GetEmailConfig().PreserveDomain,
		}
	case *mgmtv1alpha1.TransformerConfig_FirstNameConfig:
		t.FirstName = &FirstNameConfig{
			PreserveLength: tr.GetFirstNameConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_LastNameConfig:
		t.LastName = &LastNameConfig{
			PreserveLength: tr.GetLastNameConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_FullNameConfig:
		t.FullName = &FullNameConfig{
			PreserveLength: tr.GetFullNameConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_PassthroughConfig:
		t.Passthrough = &PassthroughConfig{}

	case *mgmtv1alpha1.TransformerConfig_UuidConfig:
		t.Uuid = &UuidConfig{
			IncludeHyphen: tr.GetUuidConfig().IncludeHyphen,
		}
	case *mgmtv1alpha1.TransformerConfig_PhoneNumberConfig:
		t.PhoneNumber = &PhoneNumberConfig{
			IncludeHyphens: tr.GetPhoneNumberConfig().IncludeHyphens,
			E164Format:     tr.GetPhoneNumberConfig().E164Format,
			PreserveLength: tr.GetPhoneNumberConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_IntPhoneNumberConfig:
		t.IntPhoneNumber = &IntPhoneNumberConfig{
			PreserveLength: tr.GetIntPhoneNumberConfig().PreserveLength,
		}
	case *mgmtv1alpha1.TransformerConfig_NullConfig:
		t.Null = &NullConfig{}

	case *mgmtv1alpha1.TransformerConfig_RandomStringConfig:
		t.RandomString = &RandomStringConfig{
			PreserveLength: tr.GetRandomStringConfig().PreserveLength,
			StrLength:      tr.GetRandomStringConfig().GetStrLength(),
		}
	case *mgmtv1alpha1.TransformerConfig_RandomBoolConfig:
		t.RandomBool = &RandomBoolConfig{}

	case *mgmtv1alpha1.TransformerConfig_RandomIntConfig:
		t.RandomInt = &RandomIntConfig{
			PreserveLength: tr.GetRandomIntConfig().PreserveLength,
			IntLength:      tr.GetRandomIntConfig().IntLength,
		}
	case *mgmtv1alpha1.TransformerConfig_RandomFloatConfig:
		t.RandomFloat = &RandomFloatConfig{
			PreserveLength:      tr.GetRandomFloatConfig().PreserveLength,
			DigitsBeforeDecimal: tr.GetRandomFloatConfig().DigitsBeforeDecimal,
			DigitsAfterDecimal:  tr.GetRandomFloatConfig().DigitsAfterDecimal,
		}
	case *mgmtv1alpha1.TransformerConfig_GenderConfig:
		t.Gender = &GenderConfig{
			Abbreviate: tr.GetGenderConfig().Abbreviate,
		}
	case *mgmtv1alpha1.TransformerConfig_UtcTimestampConfig:
		t.UTCTimestamp = &UTCTimestampConfig{}

	case *mgmtv1alpha1.TransformerConfig_UnixTimestampConfig:
		t.UnixTimestamp = &UnixTimestampConfig{}

	case *mgmtv1alpha1.TransformerConfig_StreetAddressConfig:
		t.StreetAddress = &StreetAddressConfig{}

	case *mgmtv1alpha1.TransformerConfig_CityConfig:
		t.City = &CityConfig{}

	case *mgmtv1alpha1.TransformerConfig_ZipcodeConfig:
		t.Zipcode = &ZipcodeConfig{}
	case *mgmtv1alpha1.TransformerConfig_StateConfig:
		t.State = &StateConfig{}

	case *mgmtv1alpha1.TransformerConfig_FullAddressConfig:
		t.FullAddress = &FullAddressConfig{}
	case *mgmtv1alpha1.TransformerConfig_CreditCardConfig:
		t.CreditCard = &CreditCardConfig{
			ValidLuhn: tr.GetCreditCardConfig().ValidLuhn,
		}
	case *mgmtv1alpha1.TransformerConfig_Sha256HashConfig:
		t.SHA256Hash = &SHA256HashConfig{}
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
	case tr.EmailConfig != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_EmailConfig{
				EmailConfig: &mgmtv1alpha1.EmailConfig{
					PreserveDomain: tr.EmailConfig.PreserveDomain,
					PreserveLength: tr.EmailConfig.PreserveLength,
				},
			},
		}
	case tr.FirstName != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_FirstNameConfig{
				FirstNameConfig: &mgmtv1alpha1.FirstName{
					PreserveLength: tr.FirstName.PreserveLength,
				},
			},
		}
	case tr.LastName != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_LastNameConfig{
				LastNameConfig: &mgmtv1alpha1.LastName{
					PreserveLength: tr.LastName.PreserveLength,
				},
			},
		}
	case tr.FullName != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_FullNameConfig{
				FullNameConfig: &mgmtv1alpha1.FullName{
					PreserveLength: tr.FullName.PreserveLength,
				},
			},
		}
	case tr.Passthrough != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
				PassthroughConfig: &mgmtv1alpha1.Passthrough{},
			},
		}
	case tr.PhoneNumber != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_PhoneNumberConfig{
				PhoneNumberConfig: &mgmtv1alpha1.PhoneNumber{
					PreserveLength: tr.PhoneNumber.PreserveLength,
					E164Format:     tr.PhoneNumber.E164Format,
					IncludeHyphens: tr.PhoneNumber.IncludeHyphens,
				},
			},
		}
	case tr.IntPhoneNumber != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_IntPhoneNumberConfig{
				IntPhoneNumberConfig: &mgmtv1alpha1.IntPhoneNumber{
					PreserveLength: tr.IntPhoneNumber.PreserveLength,
				},
			},
		}

	case tr.Uuid != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_UuidConfig{
				UuidConfig: &mgmtv1alpha1.Uuid{
					IncludeHyphen: tr.Uuid.IncludeHyphen,
				},
			},
		}
	case tr.Null != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_NullConfig{
				NullConfig: &mgmtv1alpha1.Null{},
			},
		}
	case tr.RandomString != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_RandomStringConfig{
				RandomStringConfig: &mgmtv1alpha1.RandomString{
					PreserveLength: tr.RandomString.PreserveLength,
					StrLength:      tr.RandomString.StrLength,
				},
			},
		}
	case tr.RandomBool != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_RandomBoolConfig{
				RandomBoolConfig: &mgmtv1alpha1.RandomBool{},
			},
		}
	case tr.RandomInt != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_RandomIntConfig{
				RandomIntConfig: &mgmtv1alpha1.RandomInt{
					PreserveLength: tr.RandomInt.PreserveLength,
					IntLength:      tr.RandomInt.IntLength,
				},
			},
		}
	case tr.RandomFloat != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_RandomFloatConfig{
				RandomFloatConfig: &mgmtv1alpha1.RandomFloat{
					PreserveLength:      tr.RandomFloat.PreserveLength,
					DigitsBeforeDecimal: tr.RandomFloat.DigitsBeforeDecimal,
					DigitsAfterDecimal:  tr.RandomFloat.DigitsAfterDecimal,
				},
			},
		}
	case tr.Gender != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenderConfig{
				GenderConfig: &mgmtv1alpha1.Gender{
					Abbreviate: tr.Gender.Abbreviate,
				},
			},
		}
	case tr.UTCTimestamp != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_UtcTimestampConfig{
				UtcTimestampConfig: &mgmtv1alpha1.UTCTimestamp{},
			},
		}
	case tr.UnixTimestamp != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_UnixTimestampConfig{
				UnixTimestampConfig: &mgmtv1alpha1.UnixTimestamp{},
			},
		}
	case tr.StreetAddress != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_StreetAddressConfig{
				StreetAddressConfig: &mgmtv1alpha1.StreetAddress{},
			},
		}
	case tr.City != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_CityConfig{
				CityConfig: &mgmtv1alpha1.City{},
			},
		}
	case t.Zipcode != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_ZipcodeConfig{
				ZipcodeConfig: &mgmtv1alpha1.Zipcode{},
			},
		}
	case tr.State != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_StateConfig{
				StateConfig: &mgmtv1alpha1.State{},
			},
		}
	case tr.FullAddress != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_FullAddressConfig{
				FullAddressConfig: &mgmtv1alpha1.FullAddress{},
			},
		}
	case tr.CreditCard != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_CreditCardConfig{
				CreditCardConfig: &mgmtv1alpha1.CreditCard{
					ValidLuhn: tr.CreditCard.ValidLuhn,
				},
			},
		}
	case tr.SHA256Hash != nil:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Sha256HashConfig{
				Sha256HashConfig: &mgmtv1alpha1.SHA256Hash{},
			},
		}
	default:
		return &mgmtv1alpha1.TransformerConfig{}
	}
}
