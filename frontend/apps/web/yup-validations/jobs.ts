import { SubsetFormValues } from '@/app/[account]/new/job/schema';
import { TransformerConfigSchema } from '@/app/[account]/new/transformer/schema';
import {
  AwsS3DestinationConnectionOptions,
  Connection,
  JobDestinationOptions,
  JobMappingTransformer,
  MysqlDestinationConnectionOptions,
  MysqlTruncateTableConfig,
  PostgresDestinationConnectionOptions,
  PostgresSourceSchemaOption,
  PostgresSourceTableOption,
  PostgresTruncateTableConfig,
  TransformerConfig,
} from '@neosync/sdk';
import * as Yup from 'yup';

// Yup schema form JobMappingTransformers
const JobMappingTransformerForm = Yup.object({
  source: Yup.string().required(),
  config: TransformerConfigSchema,
});

// Simplified version of a job mapping transformer for use with react-hook-form only
export type JobMappingTransformerForm = Yup.InferType<
  typeof JobMappingTransformerForm
>;

export function convertJobMappingTransformerToForm(
  jmt: JobMappingTransformer
): JobMappingTransformerForm {
  return {
    source: jmt.source,
    config: convertTransformerConfigToForm(jmt.config),
  };
}
export function convertJobMappingTransformerFormToJobMappingTransformer(
  form: JobMappingTransformerForm
): JobMappingTransformer {
  return new JobMappingTransformer({
    source: form.source,
    config: convertTransformerConfigSchemaToTransformerConfig(form.config),
  });
}

export function convertTransformerConfigToForm(
  tc?: TransformerConfig
): TransformerConfigSchema {
  const config = tc?.config ?? { case: '', value: {} };
  if (!config.case) {
    return { case: '', value: {} };
  }
  return config;
}

export function convertTransformerConfigSchemaToTransformerConfig(
  tcs: TransformerConfigSchema
): TransformerConfig {
  return TransformerConfig.fromJson({
    [tcs.case ?? '']: tcs.value,
  });
}

export type SchemaFormValues = Yup.InferType<typeof SCHEMA_FORM_SCHEMA>;

export const JOB_MAPPING_SCHEMA = Yup.object({
  schema: Yup.string().required(),
  table: Yup.string().required(),
  column: Yup.string().required(),
  dataType: Yup.string().required(),
  transformer: JobMappingTransformerForm,
}).required();
export type JobMappingFormValues = Yup.InferType<typeof JOB_MAPPING_SCHEMA>;

export const SCHEMA_FORM_SCHEMA = Yup.object({
  mappings: Yup.array().of(JOB_MAPPING_SCHEMA).required(),
});

export const SOURCE_FORM_SCHEMA = Yup.object({
  sourceId: Yup.string().uuid('Source is required').required(),
  sourceOptions: Yup.object({
    haltOnNewColumnAddition: Yup.boolean().optional(),
  }),
});

export const DESTINATION_FORM_SCHEMA = Yup.object({
  connectionId: Yup.string()
    .uuid('Destination is required')
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
          case: 'postgresOptions',
          value: new PostgresDestinationConnectionOptions({
            truncateTable: new PostgresTruncateTableConfig({
              truncateBeforeInsert:
                values.destinationOptions.truncateBeforeInsert ?? false,
              cascade: values.destinationOptions.truncateCascade ?? false,
            }),
            initTableSchema: values.destinationOptions.initTableSchema,
          }),
        },
      });
    }
    case 'mysqlConfig': {
      return new JobDestinationOptions({
        config: {
          case: 'mysqlOptions',
          value: new MysqlDestinationConnectionOptions({
            truncateTable: new MysqlTruncateTableConfig({
              truncateBeforeInsert:
                values.destinationOptions.truncateBeforeInsert ?? false,
            }),
            initTableSchema: values.destinationOptions.initTableSchema,
          }),
        },
      });
    }
    case 'awsS3Config': {
      return new JobDestinationOptions({
        config: {
          case: 'awsS3Options',
          value: new AwsS3DestinationConnectionOptions({}),
        },
      });
    }
    default: {
      return new JobDestinationOptions();
    }
  }
}

export function toPostgresSourceSchemaOptions(
  subsets: SubsetFormValues['subsets']
): PostgresSourceSchemaOption[] {
  const schemaMap = subsets.reduce(
    (map, subset) => {
      if (!map[subset.schema]) {
        map[subset.schema] = new PostgresSourceSchemaOption({
          schema: subset.schema,
          tables: [],
        });
      }
      map[subset.schema].tables.push(
        new PostgresSourceTableOption({
          table: subset.table,
          whereClause: subset.whereClause,
        })
      );
      return map;
    },
    {} as Record<string, PostgresSourceSchemaOption>
  );
  return Object.values(schemaMap);
}
