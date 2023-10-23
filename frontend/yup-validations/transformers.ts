import {
  City,
  EmailConfig,
  FirstName,
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
  RandomString_StringCase,
  StreetAddress,
  Transformer,
  TransformerConfig,
  UTCTimestamp,
  UnixTimestamp,
  Uuid,
  Zipcode,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';

interface EmailTransformer {
  value: string;
  config: EmailTransformerConfigs;
}

interface EmailTransformerConfigs {
  preserveDomain: boolean;
  preserveLength: boolean;
}

interface UuidTransformer {
  value: string;
  config: UuidTransformerConfigs;
}

interface UuidTransformerConfigs {
  includeHyphen: boolean;
}

interface FirstNameTransformer {
  value: string;
  config: FirstNameTransformerConfigs;
}

interface FirstNameTransformerConfigs {
  preserveLength: boolean;
}

interface LastNameTransformer {
  value: string;
  config: LastNameTransformerConfigs;
}

interface LastNameTransformerConfigs {
  preserveLength: boolean;
}

interface FullNameTransformer {
  value: string;
  config: FullNameTransformerConfigs;
}

interface FullNameTransformerConfigs {
  preserveLength: boolean;
}

interface PhoneNumberTransformer {
  value: string;
  config: PhoneNumberTransformerConfigs;
}

interface PhoneNumberTransformerConfigs {
  preserveLength: boolean;
  e164Format: boolean;
  includeHyphens: boolean;
}

interface IntPhoneNumberTransformer {
  value: string;
  config: IntPhoneNumberTransformerConfigs;
}

interface IntPhoneNumberTransformerConfigs {
  preserveLength: boolean;
}

interface RandomStringTransformer {
  value: string;
  config: RandomStringTransformerConfigs;
}

interface RandomStringTransformerConfigs {
  preserveLength: boolean;
  strCase: RandomString_StringCase;
  strLength: number;
}

interface RandomIntTransformer {
  value: string;
  config: RandomIntTransformerConfigs;
}

interface RandomIntTransformerConfigs {
  preserveLength: boolean;
  intLength: number;
}
interface RandomFloatTransformer {
  value: string;
  config: RandomFloatTransformerConfigs;
}

interface RandomFloatTransformerConfigs {
  preserveLength: boolean;
  digitsBeforeDecimal: number;
  digitsAftereDecimal: number;
}

interface GenderTransformer {
  value: string;
  config: GenderTransformerConfigs;
}

interface GenderTransformerConfigs {
  abbreviate: boolean;
}

export function toTransformerConfigOptions(t: {
  value: string;
  config: {};
}): Transformer {
  if (!t) {
    return new Transformer();
  }

  switch (t.value) {
    case 'email': {
      const et = t as EmailTransformer; //cast to email transformer to access fields in config object
      return new Transformer({
        value: et.value,
        config: new TransformerConfig({
          config: {
            case: 'emailConfig',
            value: new EmailConfig({
              preserveDomain: et.config.preserveDomain,
              preserveLength: et.config.preserveLength,
            }),
          },
        }),
      });
    }
    case 'passthrough': {
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'passthroughConfig',
            value: new Passthrough({}),
          },
        }),
      });
    }
    case 'uuid': {
      const ut = t as UuidTransformer;
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'uuidConfig',
            value: new Uuid({
              includeHyphen: ut.config.includeHyphen,
            }),
          },
        }),
      });
    }
    case 'first_name': {
      const ft = t as FirstNameTransformer;
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'firstNameConfig',
            value: new FirstName({
              preserveLength: ft.config.preserveLength,
            }),
          },
        }),
      });
    }
    case 'last_name': {
      const ft = t as LastNameTransformer;
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'lastNameConfig',
            value: new LastName({
              preserveLength: ft.config.preserveLength,
            }),
          },
        }),
      });
    }
    case 'full_name': {
      const ft = t as FullNameTransformer;
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'fullNameConfig',
            value: new FullName({
              preserveLength: ft.config.preserveLength,
            }),
          },
        }),
      });
    }
    case 'phone_number': {
      const pt = t as PhoneNumberTransformer;
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'phoneNumberConfig',
            value: new PhoneNumber({
              preserveLength: pt.config.preserveLength,
              e164Format: pt.config.e164Format,
              includeHyphens: pt.config.includeHyphens,
            }),
          },
        }),
      });
    }
    case 'int_phone_number': {
      const pt = t as IntPhoneNumberTransformer;
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'intPhoneNumberConfig',
            value: new IntPhoneNumber({
              preserveLength: pt.config.preserveLength,
            }),
          },
        }),
      });
    }
    case 'null': {
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'nullConfig',
            value: new Null({}),
          },
        }),
      });
    }
    case 'random_string': {
      const rs = t as RandomStringTransformer;
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'randomStringConfig',
            value: new RandomString({
              preserveLength: rs.config.preserveLength,
              strCase: rs.config.strCase,
              strLength: BigInt(rs.config.strLength ?? 10),
            }),
          },
        }),
      });
    }
    case 'random_bool': {
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'randomBoolConfig',
            value: new RandomBool({}),
          },
        }),
      });
    }
    case 'random_int': {
      const ri = t as RandomIntTransformer;
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'randomIntConfig',
            value: new RandomInt({
              preserveLength: ri.config.preserveLength,
              intLength: BigInt(ri.config.intLength ?? 4),
            }),
          },
        }),
      });
    }
    case 'random_float': {
      const rf = t as RandomFloatTransformer;
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'randomFloatConfig',
            value: new RandomFloat({
              preserveLength: rf.config.preserveLength,
              digitsBeforeDecimal: BigInt(rf.config.digitsBeforeDecimal ?? 2),
              digitsAfterDecimal: BigInt(rf.config.digitsAftereDecimal ?? 3),
            }),
          },
        }),
      });
    }
    case 'gender': {
      const g = t as GenderTransformer;
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'genderConfig',
            value: new Gender({
              abbreviate: g.config.abbreviate,
            }),
          },
        }),
      });
    }
    case 'utc_timestamp': {
      return new Transformer({
        value: t.value,
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
        value: t.value,
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
        value: t.value,
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
        value: t.value,
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
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'zipcodeConfig',
            value: new Zipcode({}),
          },
        }),
      });
    }
    default: {
      return new Transformer({
        value: t.value,
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
