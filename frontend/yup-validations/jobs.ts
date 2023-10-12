import { Connection } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import {
  EmailConfig,
  FirstName,
  JobDestinationOptions,
  Passthrough,
  PhoneNumber,
  SqlDestinationConnectionOptions,
  Transformer,
  TransformerConfig,
  TruncateTableConfig,
  Uuidv4,
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
  exclude: Yup.boolean(),
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
    case 'uuid_v4': {
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'uuidConfig',
            value: new Uuidv4({}),
          },
        }),
      });
    }
    case 'first_name': {
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'firstNameConfig',
            value: new FirstName({}),
          },
        }),
      });
    }
    case 'phone_number': {
      return new Transformer({
        value: t.value,
        config: new TransformerConfig({
          config: {
            case: 'phoneNumberConfig',
            value: new PhoneNumber({}),
          },
        }),
      });
    }
    default: {
      return new Transformer({});
    }
  }
}
