import { Connection } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import {
  JobDestinationOptions,
  SqlDestinationConnectionOptions,
  TruncateTableConfig,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import * as Yup from 'yup';

const TRANSFORMER_CONFIG = Yup.object().oneOf([
  Yup.object().shape({
    email_config: Yup.object().shape({
      preserve_domain: Yup.boolean(),
      preserve_length: Yup.boolean(),
    }),
  }),
  Yup.string(),
  //add new transformer configs to this array
]);

export type TransformerConfigSchema = Yup.InferType<typeof TRANSFORMER_CONFIG>;

const TRANSFORMER_SCHEMA = Yup.object({
  value: Yup.string().required(),
  config: TRANSFORMER_CONFIG,
});

export type TransformerSchema = Yup.InferType<typeof TRANSFORMER_SCHEMA>;

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
