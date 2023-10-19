import { Connection } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import {
  EmailConfig,
  FirstName,
  FullName,
  IntPhoneNumber,
  JobDestinationOptions,
  LastName,
  Null,
  Passthrough,
  PhoneNumber,
  RandomString,
  RandomString_StringCase,
  SqlDestinationConnectionOptions,
  Transformer,
  TransformerConfig,
  TruncateTableConfig,
  Uuid,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import * as Yup from 'yup';

const TRANSFORMER_SCHEMA = Yup.object().shape({
  value: Yup.string().required(),
  config: Yup.object()
    .shape({})
    .when('value', {
      is: 'email',
      then: () =>
        Yup.object().shape({
          preserve_domain: Yup.boolean().notRequired(),
          preserve_length: Yup.boolean().notRequired(),
        }),
    }),
});

export type SchemaFormValues = Yup.InferType<typeof SCHEMA_FORM_SCHEMA>;

const JOB_MAPPING_SCHEMA = Yup.object({
  schema: Yup.string().required(),
  table: Yup.string().required(),
  column: Yup.string().required(),
  dataType: Yup.string().required(),
  transformer: TRANSFORMER_SCHEMA,
}).required();
export type JobMappingFormValues = Yup.InferType<typeof JOB_MAPPING_SCHEMA>;

export const SCHEMA_FORM_SCHEMA = Yup.object({
  mappings: Yup.array().of(JOB_MAPPING_SCHEMA).required(),
});

export const SOURCE_FORM_SCHEMA = Yup.object({
  sourceId: Yup.string().uuid('source is required').required(),
  sourceOptions: Yup.object({
    haltOnNewColumnAddition: Yup.boolean().optional(),
  }),
});

export const DESTINATION_FORM_SCHEMA = Yup.object({
  connectionId: Yup.string()
    .uuid('destination is required')
    .required()
    .test(
      'checkConnectionUnique',
      'Destination must be different from source.',
      (value, ctx) => {
        if (ctx.from) {
          const { sourceId } = ctx.from[ctx.from.length - 1].value;
          if (value == sourceId) {
            return false;
          }
        }

        return true;
      }
    ),
  destinationOptions: Yup.object({
    truncateBeforeInsert: Yup.boolean().optional(),
    truncateCascade: Yup.boolean().optional(),
    initTableSchema: Yup.boolean().optional(),
  }),
}).required();
type DestinationFormValues = Yup.InferType<typeof DESTINATION_FORM_SCHEMA>;

export function toJobDestinationOptions(
  values: DestinationFormValues,
  connection?: Connection
): JobDestinationOptions {
  if (!connection) {
    return new JobDestinationOptions();
  }
  switch (connection.connectionConfig?.config.case) {
    case 'pgConfig': {
      return new JobDestinationOptions({
        config: {
          case: 'sqlOptions',
          value: new SqlDestinationConnectionOptions({
            truncateTable: new TruncateTableConfig({
              truncateBeforeInsert:
                values.destinationOptions.truncateBeforeInsert ?? false,
              cascade: values.destinationOptions.truncateCascade ?? false,
            }),
            initTableSchema: values.destinationOptions.initTableSchema,
          }),
        },
      });
    }
    default: {
      return new JobDestinationOptions();
    }
  }
}

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
    case 'generic_string': {
      const gs = t as RandomStringTransformer;
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'randomStringConfig',
            value: new RandomString({
              preserveLength: gs.config.preserveLength,
              strCase: gs.config.strCase,
              strLength: BigInt(gs.config.strLength ?? 10),
            }),
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
