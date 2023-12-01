import {
  CustomTransformer,
  GenerateBool,
  GenerateCardNumber,
  GenerateCity,
  GenerateE164Number,
  GenerateEmail,
  GenerateFirstName,
  GenerateFloat,
  GenerateFullAddress,
  GenerateFullName,
  GenerateGender,
  GenerateInt,
  GenerateInt64Phone,
  GenerateLastName,
  GenerateRealisticEmail,
  GenerateSSN,
  GenerateSha256Hash,
  GenerateState,
  GenerateStreetAddress,
  GenerateString,
  GenerateStringPhone,
  GenerateUnixTimestamp,
  GenerateUsername,
  GenerateUtcTimestamp,
  GenerateUuid,
  GenerateZipcode,
  Null,
  Passthrough,
  TransformE164Phone,
  TransformEmail,
  TransformFirstName,
  TransformFloat,
  TransformFullName,
  TransformInt,
  TransformIntPhone,
  TransformLastName,
  TransformPhone,
  TransformString,
  Transformer,
  TransformerConfig,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';

interface TransformEmailTransformer {
  case?: string | undefined;
  value: TransformEmailConfigs;
}

interface TransformEmailConfigs {
  preserveDomain: boolean;
  preserveLength: boolean;
}

interface GenerateCardNumberTransformer {
  case?: string | undefined;
  value: GenerateCardNumberConfig;
}

interface GenerateCardNumberConfig {
  validLuhn: boolean;
}

interface GenerateE164NumberTransformer {
  case?: string | undefined;
  value: GenerateE164NumberConfig;
}

interface GenerateE164NumberConfig {
  length: number;
}

interface GenerateFloatTransformer {
  case?: string | undefined;
  value: GenerateFloatTransformerConfig;
}

interface GenerateFloatTransformerConfig {
  sign: string;
  digitsBeforeDecimal: number;
  digitsAfterDecimal: number;
}

interface GenerateGenderTransformer {
  case?: string | undefined;
  value: GenerateGenderConfig;
}

interface GenerateGenderConfig {
  abbreviate: boolean;
}

interface GenerateIntTransformer {
  case?: string | undefined;
  value: GenerateIntConfig;
}

interface GenerateIntConfig {
  length: number;
  sign: string;
}

interface GenerateStringPhoneTransformer {
  case?: string | undefined;
  value: GenerateStringPhoneConfig;
}

interface GenerateStringPhoneConfig {
  includeHyphens: boolean;
}

interface GenerateStringTransformer {
  case?: string | undefined;
  value: GenerateStringConfig;
}

interface GenerateStringConfig {
  length: number;
}

interface GenerateUuidTransformer {
  case?: string | undefined;
  value: GenerateUuidConfig;
}

interface GenerateUuidConfig {
  includeHyphens: boolean;
}

interface TransformE164PhoneTransformer {
  case?: string | undefined;
  value: TransformE164PhoneConfig;
}

interface TransformE164PhoneConfig {
  preserveLength: boolean;
}

interface TransformFirstNameTransformer {
  case?: string | undefined;
  value: TransformFirstNameConfig;
}

interface TransformFirstNameConfig {
  preserveLength: boolean;
}

interface TransformFloatTransformer {
  case?: string | undefined;
  value: TransformFloatConfig;
}

interface TransformFloatConfig {
  preserveLength: boolean;
  preserveSign: boolean;
}

interface TransformFullNameTransformer {
  case?: string | undefined;
  value: TransformFullNameConfig;
}

interface TransformFullNameConfig {
  preserveLength: boolean;
}

interface TransformIntPhoneTransformer {
  case?: string | undefined;
  value: TransformerIntPhoneConfig;
}

interface TransformerIntPhoneConfig {
  preserveLength: boolean;
}

interface TransformIntTransformer {
  case?: string | undefined;
  value: TransformIntConfig;
}

interface TransformIntConfig {
  preserveLength: boolean;
  preserveSign: boolean;
}

interface TransformLastNameTransformer {
  case?: string | undefined;
  value: TransformerLastNameConfig;
}

interface TransformerLastNameConfig {
  preserveLength: boolean;
}

interface TransformPhoneTransformer {
  case?: string | undefined;
  value: TransformPhoneConfig;
}

interface TransformPhoneConfig {
  preserveLength: boolean;
  IncludeHyphens: boolean;
}

interface TransformStringTransformer {
  case?: string | undefined;
  value: TransformStringConfig;
}

interface TransformStringConfig {
  preserveLength: boolean;
}

export function ToTransformerConfigOptions(
  t: {
    value: string;
    config: { config: { case?: string | undefined; value: {} } };
  },
  merged: CustomTransformer[]
): Transformer {
  const val = merged.find((item) => item.name.toLowerCase() == t.value);

  if (!t) {
    return new Transformer();
  }

  switch (val?.source) {
    case 'generate_email': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateEmailConfig',
            value: new GenerateEmail({}),
          },
        }),
      });
    }
    case 'generate_realistic_email': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateRealisticEmailConfig',
            value: new GenerateRealisticEmail({}),
          },
        }),
      });
    }
    case 'transform_email': {
      const te = t.config.config as TransformEmailTransformer; //cast to email transformer to access fields in config object
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'transformEmailConfig',
            value: new TransformEmail({
              preserveDomain: te.value.preserveDomain,
              preserveLength: te.value.preserveLength,
            }),
          },
        }),
      });
    }
    case 'generate_bool': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateBoolConfig',
            value: new GenerateBool({}),
          },
        }),
      });
    }
    case 'generate_card_number': {
      const gcn = t.config.config as GenerateCardNumberTransformer; //cast to email transformer to access fields in config object
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateCardNumberConfig',
            value: new GenerateCardNumber({
              validLuhn: gcn.value.validLuhn,
            }),
          },
        }),
      });
    }
    case 'generate_city': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateCityConfig',
            value: new GenerateCity({}),
          },
        }),
      });
    }
    case 'generate_e164_number': {
      const gen = t.config.config as GenerateE164NumberTransformer; //cast to email transformer to access fields in config object
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateE164NumberConfig',
            value: new GenerateE164Number({
              length: BigInt(gen.value.length),
            }),
          },
        }),
      });
    }
    case 'generate_first_name': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateFirstNameConfig',
            value: new GenerateFirstName({}),
          },
        }),
      });
    }
    case 'generate_float': {
      const gt = t.config.config as GenerateFloatTransformer; //cast to email transformer to access fields in config object
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateFloatConfig',
            value: new GenerateFloat({
              sign: gt.value.sign,
              digitsBeforeDecimal: BigInt(gt.value.digitsBeforeDecimal),
              digitsAfterDecimal: BigInt(gt.value.digitsAfterDecimal),
            }),
          },
        }),
      });
    }
    case 'generate_full_address': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateFullAddressConfig',
            value: new GenerateFullAddress({}),
          },
        }),
      });
    }
    case 'generate_full_name': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateFullNameConfig',
            value: new GenerateFullName({}),
          },
        }),
      });
    }
    case 'generate_gender': {
      const gg = t.config.config as GenerateGenderTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateGenderConfig',
            value: new GenerateGender({
              abbreviate: gg.value.abbreviate,
            }),
          },
        }),
      });
    }
    case 'generate_int64_phone': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateInt64PhoneConfig',
            value: new GenerateInt64Phone({}),
          },
        }),
      });
    }
    case 'generate_int': {
      const gi = t.config.config as GenerateIntTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateIntConfig',
            value: new GenerateInt({
              length: BigInt(gi.value.length),
              sign: gi.value.sign,
            }),
          },
        }),
      });
    }
    case 'generate_last_name': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateLastNameConfig',
            value: new GenerateLastName({}),
          },
        }),
      });
    }
    case 'generate_sha256hash': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateSha256hashConfig',
            value: new GenerateSha256Hash({}),
          },
        }),
      });
    }
    case 'generate_ssn': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateSsnConfig',
            value: new GenerateSSN({}),
          },
        }),
      });
    }
    case 'generate_state': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateStateConfig',
            value: new GenerateState({}),
          },
        }),
      });
    }
    case 'generate_street_address': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateStreetAddressConfig',
            value: new GenerateStreetAddress({}),
          },
        }),
      });
    }
    case 'generate_string_phone': {
      const gsp = t.config.config as GenerateStringPhoneTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateStringPhoneConfig',
            value: new GenerateStringPhone({
              includeHyphens: gsp.value.includeHyphens,
            }),
          },
        }),
      });
    }
    case 'generate_string': {
      const gs = t.config.config as GenerateStringTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateStringConfig',
            value: new GenerateString({
              length: BigInt(gs.value.length),
            }),
          },
        }),
      });
    }
    case 'generate_unixtimestamp': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateUnixtimestampConfig',
            value: new GenerateUnixTimestamp({}),
          },
        }),
      });
    }
    case 'generate_username': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateUsernameConfig',
            value: new GenerateUsername({}),
          },
        }),
      });
    }
    case 'generate_utctimestamp': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateUtctimestampConfig',
            value: new GenerateUtcTimestamp({}),
          },
        }),
      });
    }
    case 'generate_uuid': {
      const gu = t.config.config as GenerateUuidTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateUuidConfig',
            value: new GenerateUuid({
              includeHyphens: gu.value.includeHyphens,
            }),
          },
        }),
      });
    }
    case 'generate_zipcode': {
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'generateZipcodeConfig',
            value: new GenerateZipcode({}),
          },
        }),
      });
    }
    case 'transform_e164_phone': {
      const te = t.config.config as TransformE164PhoneTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'transformE164PhoneConfig',
            value: new TransformE164Phone({
              preserveLength: te.value.preserveLength,
            }),
          },
        }),
      });
    }
    case 'transform_first_name': {
      const tf = t.config.config as TransformFirstNameTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'transformFirstNameConfig',
            value: new TransformFirstName({
              preserveLength: tf.value.preserveLength,
            }),
          },
        }),
      });
    }
    case 'transform_float': {
      const tf = t.config.config as TransformFloatTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'transformFloatConfig',
            value: new TransformFloat({
              preserveSign: tf.value.preserveSign,
              preserveLength: tf.value.preserveLength,
            }),
          },
        }),
      });
    }
    case 'transform_full_name': {
      const tf = t.config.config as TransformFullNameTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'transformFullNameConfig',
            value: new TransformFullName({
              preserveLength: tf.value.preserveLength,
            }),
          },
        }),
      });
    }
    case 'transform_int_phone': {
      const ti = t.config.config as TransformIntPhoneTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'transformIntPhoneConfig',
            value: new TransformIntPhone({
              preserveLength: ti.value.preserveLength,
            }),
          },
        }),
      });
    }
    case 'transform_int': {
      const ti = t.config.config as TransformIntTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'transformIntConfig',
            value: new TransformInt({
              preserveLength: ti.value.preserveLength,
              preserveSign: ti.value.preserveSign,
            }),
          },
        }),
      });
    }
    case 'transform_last_name': {
      const tl = t.config.config as TransformLastNameTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'transformLastNameConfig',
            value: new TransformLastName({
              preserveLength: tl.value.preserveLength,
            }),
          },
        }),
      });
    }
    case 'transform_phone': {
      const tp = t.config.config as TransformPhoneTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'transformPhoneConfig',
            value: new TransformPhone({
              preserveLength: tp.value.preserveLength,
              includeHyphens: tp.value.IncludeHyphens,
            }),
          },
        }),
      });
    }
    case 'transform_string': {
      const ts = t.config.config as TransformStringTransformer;
      return new Transformer({
        value: val.source,
        config: new TransformerConfig({
          config: {
            case: 'transformStringConfig',
            value: new TransformString({
              preserveLength: ts.value.preserveLength,
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
            case: 'nullconfig',
            value: new Null({}),
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
