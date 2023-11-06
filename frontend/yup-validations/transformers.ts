import {
  City,
  CreditCard,
  CustomTransformer,
  EmailConfig,
  FirstName,
  FullAddress,
  FullName,
  Gender,
  IntPhoneNumber,
  LastName,
  Null,
  Passthrough,
  PhoneNumber,
  RandomBool,
  RandomFloat,
  RandomInt,
  RandomString,
  SHA256Hash,
  State,
  StreetAddress,
  Transformer,
  TransformerConfig,
  UTCTimestamp,
  UnixTimestamp,
  Uuid,
  Zipcode,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';

interface EmailTransformer {
  case?: string | undefined;
  value: EmailTransformerConfigs;
}

interface EmailTransformerConfigs {
  preserveDomain: boolean;
  preserveLength: boolean;
}

interface UuidTransformer {
  case?: string | undefined;
  value: UuidTransformerConfigs;
}

interface UuidTransformerConfigs {
  includeHyphen: boolean;
}

interface FirstNameTransformer {
  case?: string | undefined;
  value: FirstNameTransformerConfigs;
}

interface FirstNameTransformerConfigs {
  preserveLength: boolean;
}

interface LastNameTransformer {
  case?: string | undefined;
  value: LastNameTransformerConfigs;
}

interface LastNameTransformerConfigs {
  preserveLength: boolean;
}

interface FullNameTransformer {
  case?: string | undefined;
  value: FullNameTransformerConfigs;
}

interface FullNameTransformerConfigs {
  preserveLength: boolean;
}

interface PhoneNumberTransformer {
  case?: string | undefined;
  value: PhoneNumberTransformerConfigs;
}

interface PhoneNumberTransformerConfigs {
  preserveLength: boolean;
  e164Format: boolean;
  includeHyphens: boolean;
}

interface IntPhoneNumberTransformer {
  case?: string | undefined;
  value: IntPhoneNumberTransformerConfigs;
}

interface IntPhoneNumberTransformerConfigs {
  preserveLength: boolean;
}

interface RandomStringTransformer {
  case?: string | undefined;
  value: RandomStringTransformerConfigs;
}

interface RandomStringTransformerConfigs {
  preserveLength: boolean;
  strLength: number;
}

interface RandomIntTransformer {
  case?: string | undefined;
  value: RandomIntTransformerConfigs;
}

interface RandomIntTransformerConfigs {
  preserveLength: boolean;
  intLength: number;
}

interface RandomFloatTransformer {
  case?: string | undefined;
  value: RandomFloatTransformerConfigs;
}

interface RandomFloatTransformerConfigs {
  preserveLength: boolean;
  digitsBeforeDecimal: number;
  digitsAftereDecimal: number;
}

interface GenderTransformer {
  case?: string | undefined;
  value: GenderTransformerConfigs;
}

interface GenderTransformerConfigs {
  abbreviate: boolean;
}

interface CreditCardTransformer {
  case?: string | undefined;
  value: CreditCardTransformerConfigs;
}

interface CreditCardTransformerConfigs {
  validLuhn: boolean;
}

export function ToTransformerConfigOptions(
  t: {
    value: string;
    config: { config: { case?: string | undefined; value: {} } };
  },
  merged: CustomTransformer[]
): Transformer {
  const val = merged.find((item) => item.name == t.value);

  if (!t) {
    return new Transformer();
  }

  switch (val?.source) {
    case 'email': {
      const et = t.config.config as EmailTransformer; //cast to email transformer to access fields in config object
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'emailConfig',
            value: new EmailConfig({
              preserveDomain: et.value.preserveDomain,
              preserveLength: et.value.preserveLength,
            }),
          },
        }),
      });
    }
    case 'passthrough': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'passthroughConfig',
            value: new Passthrough({}),
          },
        }),
      });
    }
    case 'uuid': {
      const ut = t.config.config as UuidTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'uuidConfig',
            value: new Uuid({
              includeHyphen: ut.value.includeHyphen,
            }),
          },
        }),
      });
    }
    case 'first_name': {
      const ft = t.config.config as FirstNameTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'firstNameConfig',
            value: new FirstName({
              preserveLength: ft.value.preserveLength,
            }),
          },
        }),
      });
    }
    case 'last_name': {
      const ft = t.config.config as LastNameTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'lastNameConfig',
            value: new LastName({
              preserveLength: ft.value.preserveLength,
            }),
          },
        }),
      });
    }
    case 'full_name': {
      const ft = t.config.config as FullNameTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'fullNameConfig',
            value: new FullName({
              preserveLength: ft.value.preserveLength,
            }),
          },
        }),
      });
    }
    case 'phone_number': {
      const pt = t.config.config as PhoneNumberTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'phoneNumberConfig',
            value: new PhoneNumber({
              preserveLength: pt.value.preserveLength,
              e164Format: pt.value.e164Format,
              includeHyphens: pt.value.includeHyphens,
            }),
          },
        }),
      });
    }
    case 'int_phone_number': {
      const pt = t.config.config as IntPhoneNumberTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'intPhoneNumberConfig',
            value: new IntPhoneNumber({
              preserveLength: pt.value.preserveLength,
            }),
          },
        }),
      });
    }
    case 'null': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'nullConfig',
            value: new Null({}),
          },
        }),
      });
    }
    case 'random_string': {
      const rs = t.config.config as RandomStringTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'randomStringConfig',
            value: new RandomString({
              preserveLength: rs.value.preserveLength,
              strLength: BigInt(rs.value.strLength ?? 10),
            }),
          },
        }),
      });
    }
    case 'random_bool': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'randomBoolConfig',
            value: new RandomBool({}),
          },
        }),
      });
    }
    case 'random_int': {
      const ri = t.config.config as RandomIntTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'randomIntConfig',
            value: new RandomInt({
              preserveLength: ri.value.preserveLength,
              intLength: BigInt(ri.value.intLength ?? 4),
            }),
          },
        }),
      });
    }
    case 'random_float': {
      const rf = t.config.config as RandomFloatTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'randomFloatConfig',
            value: new RandomFloat({
              preserveLength: rf.value.preserveLength,
              digitsBeforeDecimal: BigInt(rf.value.digitsBeforeDecimal ?? 2),
              digitsAfterDecimal: BigInt(rf.value.digitsAftereDecimal ?? 3),
            }),
          },
        }),
      });
    }
    case 'gender': {
      const g = t.config.config as GenderTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'genderConfig',
            value: new Gender({
              abbreviate: g.value.abbreviate,
            }),
          },
        }),
      });
    }
    case 'utc_timestamp': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'utcTimestampConfig',
            value: new UTCTimestamp({}),
          },
        }),
      });
    }
    case 'unix_timestamp': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'unixTimestampConfig',
            value: new UnixTimestamp({}),
          },
        }),
      });
    }
    case 'street_address': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'streetAddressConfig',
            value: new StreetAddress({}),
          },
        }),
      });
    }
    case 'city': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'cityConfig',
            value: new City({}),
          },
        }),
      });
    }
    case 'zipcode': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'zipcodeConfig',
            value: new Zipcode({}),
          },
        }),
      });
    }
    case 'state': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'stateConfig',
            value: new State({}),
          },
        }),
      });
    }
    case 'full_address': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'fullAddressConfig',
            value: new FullAddress({}),
          },
        }),
      });
    }
    case 'creditcard': {
      const g = t.config.config as CreditCardTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'creditCardConfig',
            value: new CreditCard({
              validLuhn: g.value.validLuhn,
            }),
          },
        }),
      });
    }
    case 'sha256_hash': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'sha256hashConfig',
            value: new SHA256Hash({}),
          },
        }),
      });
    }
    default: {
      return new Transformer({
        value: 'passthrough',
        config: new TransformerConfig({
          config: {
            case: 'passthroughConfig',
            value: new Passthrough({}),
          },
        }),
      });
    }
  }
}
