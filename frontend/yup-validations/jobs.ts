import { SubsetFormValues } from '@/app/new/job/schema';
import { transformerConfig } from '@/app/new/transformer/schema';
import { Connection } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import {
  JobDestinationOptions,
  MysqlDestinationConnectionOptions,
  MysqlTruncateTableConfig,
  PostgresDestinationConnectionOptions,
  PostgresSourceSchemaOption,
  PostgresSourceTableOption,
  PostgresTruncateTableConfig,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import * as Yup from 'yup';

const TRANSFORMER_SCHEMA = Yup.object().shape({
  source: Yup.string().required(),
  name: Yup.string(),
  config: transformerConfig,
});

export type SchemaFormValues = Yup.InferType<typeof SCHEMA_FORM_SCHEMA>;

export const JOB_MAPPING_COLUMN_SCHEMA = Yup.object({
  column: Yup.string().required(),
  dataType: Yup.string().required(),
  transformer: TRANSFORMER_SCHEMA,
});
const JOB_MAPPING_SCHEMA = JOB_MAPPING_COLUMN_SCHEMA.concat(
  Yup.object({
    schema: Yup.string().required(),
    table: Yup.string().required(),
  })
).required();
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
