package jsonmodels

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type Transformer struct {
	Value  string
	Config *TransformerConfigs
}

type TransformerConfigs struct {
	EmailConfig    *EmailConfigs
	FirstName      *FirstNameConfig
	LastName       *LastNameConfig
	FullName       *FullNameConfig
	Uuid           *UuidConfig
	PhoneNumber    *PhoneNumberConfig
	IntPhoneNumber *IntPhoneNumberConfig
	Passthrough    *PassthroughConfig
	Null           *NullConfig
	RandomString   *RandomStringConfig
	RandomBool     *RandomBoolConfig
	RandomInt      *RandomIntConfig
	RandomFloat    *RandomFloatConfig
	Gender         *GenderConfig
	UTCTimestamp   *UTCTimestampConfig
}

type EmailConfigs struct {
	PreserveLength bool
	PreserveDomain bool
}

type FirstNameConfig struct {
	PreserveLength bool
}

type LastNameConfig struct {
	PreserveLength bool
}

type FullNameConfig struct {
	PreserveLength bool
}
type UuidConfig struct {
	IncludeHyphen bool
}
type PhoneNumberConfig struct {
	IncludeHyphens bool
	E164Format     bool
	PreserveLength bool
}

type IntPhoneNumberConfig struct {
	PreserveLength bool
}
type PassthroughConfig struct {
}

type NullConfig struct{}

type RandomStringConfig struct {
	PreserveLength bool
	StrLength      int64
	StrCase        string
}

type RandomBoolConfig struct{}

type RandomIntConfig struct {
	PreserveLength bool
	IntLength      int64
}

type RandomFloatConfig struct {
	PreserveLength      bool
	DigitsBeforeDecimal int64
	DigitsAfterDecimal  int64
}

type GenderConfig struct {
	Abbreviate bool
}

type UTCTimestampConfig struct{}

// from API -> DB
func (t *Transformer) FromDto(tr *mgmtv1alpha1.Transformer) error {

	switch tr.Config.Config.(type) {
	case *mgmtv1alpha1.TransformerConfig_EmailConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			EmailConfig: &EmailConfigs{
				PreserveLength: tr.Config.GetEmailConfig().PreserveLength,
				PreserveDomain: tr.Config.GetEmailConfig().PreserveDomain,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_FirstNameConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			FirstName: &FirstNameConfig{
				PreserveLength: tr.Config.GetFirstNameConfig().PreserveLength,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_LastNameConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			LastName: &LastNameConfig{
				PreserveLength: tr.Config.GetLastNameConfig().PreserveLength,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_FullNameConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			FullName: &FullNameConfig{
				PreserveLength: tr.Config.GetFullNameConfig().PreserveLength,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_PassthroughConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			Passthrough: &PassthroughConfig{},
		}
	case *mgmtv1alpha1.TransformerConfig_UuidConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			Uuid: &UuidConfig{
				IncludeHyphen: tr.Config.GetUuidConfig().IncludeHyphen,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_PhoneNumberConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			PhoneNumber: &PhoneNumberConfig{
				IncludeHyphens: tr.Config.GetPhoneNumberConfig().IncludeHyphens,
				E164Format:     tr.Config.GetPhoneNumberConfig().E164Format,
				PreserveLength: tr.Config.GetPhoneNumberConfig().PreserveLength,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_IntPhoneNumberConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			IntPhoneNumber: &IntPhoneNumberConfig{
				PreserveLength: tr.Config.GetIntPhoneNumberConfig().PreserveLength,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_NullConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			Null: &NullConfig{},
		}
	case *mgmtv1alpha1.TransformerConfig_RandomStringConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			RandomString: &RandomStringConfig{
				PreserveLength: tr.Config.GetRandomStringConfig().PreserveLength,
				StrLength:      tr.Config.GetRandomStringConfig().GetStrLength(),
				StrCase:        tr.Config.GetRandomStringConfig().StrCase.String(),
			},
		}
	case *mgmtv1alpha1.TransformerConfig_RandomBoolConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			RandomBool: &RandomBoolConfig{},
		}
	case *mgmtv1alpha1.TransformerConfig_RandomIntConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			RandomInt: &RandomIntConfig{
				PreserveLength: tr.Config.GetRandomIntConfig().PreserveLength,
				IntLength:      tr.Config.GetRandomIntConfig().IntLength,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_RandomFloatConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			RandomFloat: &RandomFloatConfig{
				PreserveLength:      tr.Config.GetRandomFloatConfig().PreserveLength,
				DigitsBeforeDecimal: tr.Config.GetRandomFloatConfig().DigitsBeforeDecimal,
				DigitsAfterDecimal:  tr.Config.GetRandomFloatConfig().DigitsAfterDecimal,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_GenderConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			Gender: &GenderConfig{
				Abbreviate: tr.Config.GetGenderConfig().Abbreviate,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_UtcTimestampConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			UTCTimestamp: &UTCTimestampConfig{},
		}
	default:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{}
	}

	return nil
}

// DB -> API
func (t *Transformer) ToDto() *mgmtv1alpha1.Transformer {

	switch {
	case t.Config.EmailConfig != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_EmailConfig{
					EmailConfig: &mgmtv1alpha1.EmailConfig{
						PreserveDomain: t.Config.EmailConfig.PreserveDomain,
						PreserveLength: t.Config.EmailConfig.PreserveLength,
					},
				},
			},
		}
	case t.Config.FirstName != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_FirstNameConfig{
					FirstNameConfig: &mgmtv1alpha1.FirstName{
						PreserveLength: t.Config.FirstName.PreserveLength,
					},
				},
			},
		}
	case t.Config.LastName != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_LastNameConfig{
					LastNameConfig: &mgmtv1alpha1.LastName{
						PreserveLength: t.Config.LastName.PreserveLength,
					},
				},
			},
		}
	case t.Config.FullName != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_FullNameConfig{
					FullNameConfig: &mgmtv1alpha1.FullName{
						PreserveLength: t.Config.FullName.PreserveLength,
					},
				},
			},
		}
	case t.Config.Passthrough != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
					PassthroughConfig: &mgmtv1alpha1.Passthrough{},
				},
			},
		}
	case t.Config.PhoneNumber != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_PhoneNumberConfig{
					PhoneNumberConfig: &mgmtv1alpha1.PhoneNumber{
						PreserveLength: t.Config.PhoneNumber.PreserveLength,
						E164Format:     t.Config.PhoneNumber.E164Format,
						IncludeHyphens: t.Config.PhoneNumber.IncludeHyphens,
					},
				},
			},
		}
	case t.Config.IntPhoneNumber != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_IntPhoneNumberConfig{
					IntPhoneNumberConfig: &mgmtv1alpha1.IntPhoneNumber{
						PreserveLength: t.Config.IntPhoneNumber.PreserveLength,
					},
				},
			},
		}
	case t.Config.Uuid != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_UuidConfig{
					UuidConfig: &mgmtv1alpha1.Uuid{
						IncludeHyphen: t.Config.Uuid.IncludeHyphen,
					},
				},
			},
		}
	case t.Config.Null != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_NullConfig{
					NullConfig: &mgmtv1alpha1.Null{},
				},
			},
		}
	case t.Config.RandomString != nil:

		strCase, err := StrCaseFromString(t.Config.RandomString.StrCase)
		if err != nil {
			return &mgmtv1alpha1.Transformer{Value: t.Value}
		}

		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_RandomStringConfig{
					RandomStringConfig: &mgmtv1alpha1.RandomString{
						PreserveLength: t.Config.RandomString.PreserveLength,
						StrLength:      t.Config.RandomString.StrLength,
						StrCase:        strCase,
					},
				},
			},
		}
	case t.Config.RandomBool != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_RandomBoolConfig{
					RandomBoolConfig: &mgmtv1alpha1.RandomBool{},
				},
			},
		}
	case t.Config.RandomInt != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_RandomIntConfig{
					RandomIntConfig: &mgmtv1alpha1.RandomInt{
						PreserveLength: t.Config.RandomInt.PreserveLength,
						IntLength:      t.Config.RandomInt.IntLength,
					},
				},
			},
		}
	case t.Config.RandomFloat != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_RandomFloatConfig{
					RandomFloatConfig: &mgmtv1alpha1.RandomFloat{
						PreserveLength:      t.Config.RandomFloat.PreserveLength,
						DigitsBeforeDecimal: t.Config.RandomFloat.DigitsBeforeDecimal,
						DigitsAfterDecimal:  t.Config.RandomFloat.DigitsAfterDecimal,
					},
				},
			},
		}
	case t.Config.Gender != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenderConfig{
					GenderConfig: &mgmtv1alpha1.Gender{
						Abbreviate: t.Config.Gender.Abbreviate,
					},
				},
			},
		}
	case t.Config.UTCTimestamp != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_UtcTimestampConfig{
					UtcTimestampConfig: &mgmtv1alpha1.UTCTimestamp{},
				},
			},
		}
	default:
		return &mgmtv1alpha1.Transformer{Value: t.Value}
	}
}

func StrCaseFromString(strCase string) (mgmtv1alpha1.RandomString_StringCase, error) {
	switch strCase {
	case "UPPER":
		return mgmtv1alpha1.RandomString_STRING_CASE_UPPER, nil
	case "LOWER":
		return mgmtv1alpha1.RandomString_STRING_CASE_LOWER, nil
	case "TITLE":
		return mgmtv1alpha1.RandomString_STRING_CASE_TITLE, nil
	default:
		return mgmtv1alpha1.RandomString_STRING_CASE_LOWER, fmt.Errorf("invalid string case: %s", strCase)
	}
}
