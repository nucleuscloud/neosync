import { TransformerConfigFormValue } from '@/yup-validations/transformer-validations';
import { create, fromJson, JsonObject, toJson } from '@bufbuild/protobuf';
import {
  JobMappingTransformer,
  JobMappingTransformerSchema,
  MssqlSourceConnectionOptions_ColumnRemovalStrategy,
  MssqlSourceConnectionOptions_ColumnRemovalStrategy_HaltJobSchema,
  MssqlSourceConnectionOptions_ColumnRemovalStrategySchema,
  MysqlSourceConnectionOptions_ColumnRemovalStrategy,
  MysqlSourceConnectionOptions_ColumnRemovalStrategy_HaltJobSchema,
  MysqlSourceConnectionOptions_ColumnRemovalStrategySchema,
  PostgresSourceConnectionOptions_ColumnRemovalStrategy,
  PostgresSourceConnectionOptions_ColumnRemovalStrategy_HaltJobSchema,
  PostgresSourceConnectionOptions_ColumnRemovalStrategySchema,
  PostgresSourceConnectionOptions_NewColumnAdditionStrategy,
  PostgresSourceConnectionOptions_NewColumnAdditionStrategy_AutoMapSchema,
  PostgresSourceConnectionOptions_NewColumnAdditionStrategy_HaltJobSchema,
  PostgresSourceConnectionOptions_NewColumnAdditionStrategySchema,
  TransformerConfig,
  TransformerConfigSchema,
} from '@neosync/sdk';
import * as Yup from 'yup';
import { getDurationValidateFn } from './number';

// Yup schema form JobMappingTransformers
export const JobMappingTransformerForm = Yup.object({
  config: TransformerConfigFormValue,
});

// Simplified version of a job mapping transformer for use with react-hook-form only
export type JobMappingTransformerForm = Yup.InferType<
  typeof JobMappingTransformerForm
>;

export function convertJobMappingTransformerToForm(
  jmt: JobMappingTransformer
): JobMappingTransformerForm {
  return {
    config: convertTransformerConfigToForm(jmt.config),
  };
}
export function convertJobMappingTransformerFormToJobMappingTransformer(
  form: JobMappingTransformerForm
): JobMappingTransformer {
  return create(JobMappingTransformerSchema, {
    config: convertTransformerConfigSchemaToTransformerConfig(form.config),
  });
}

export function convertTransformerConfigToForm(
  tc?: TransformerConfig
): TransformerConfigFormValue {
  const config = tc?.config ?? { case: '', value: {} };
  if (!tc || !config.case) {
    return { case: '', value: {} };
  }
  const json = toJson(TransformerConfigSchema, tc);
  return {
    case: config.case,
    value: (json as JsonObject)[config.case],
  };
}

export function convertTransformerConfigSchemaToTransformerConfig(
  tcs: TransformerConfigFormValue
): TransformerConfig {
  if (tcs.case) {
    return fromJson(TransformerConfigSchema, { [tcs.case]: tcs.value });
  }
  return create(TransformerConfigSchema);
}

const BatchFormValues = Yup.object({
  count: Yup.number().min(0, 'Must be greater than or equal to 0').optional(),
  period: Yup.string()
    .optional()
    .test('duration', 'Must be a valid duration', getDurationValidateFn()),
});

export const JobMappingFormValues = Yup.object({
  schema: Yup.string().required('A schema is required'),
  table: Yup.string().required('A table is required'),
  column: Yup.string().required('A column is required'),
  transformer: JobMappingTransformerForm,
}).required('Job mapping values are required.');
export type JobMappingFormValues = Yup.InferType<typeof JobMappingFormValues>;

const VIRTUAL_FOREIGN_KEY_SCHEMA = Yup.object({
  schema: Yup.string().required('A schema is required'),
  table: Yup.string().required('A table is required'),
  columns: Yup.array().of(Yup.string().required('Columns are required')),
}).required('Virtural Foreign Key mappings are required.');

const VirtualForeignConstraintFormValues = Yup.object({
  schema: Yup.string().required('A schema is required'),
  table: Yup.string().required('A table is required'),
  columns: Yup.array().of(Yup.string().required('A column is required')),
  foreignKey: VIRTUAL_FOREIGN_KEY_SCHEMA,
}).required('Virtual Foreign Key Constraint values are required.');
export type VirtualForeignConstraintFormValues = Yup.InferType<
  typeof VirtualForeignConstraintFormValues
>;

export type NewColumnAdditionStrategy = 'continue' | 'halt' | 'automap';
export type ColumnRemovalStrategy = 'halt' | 'continue';

export const PostgresSourceOptionsFormValues = Yup.object({
  newColumnAdditionStrategy: Yup.string<NewColumnAdditionStrategy>()
    .oneOf(['continue', 'halt', 'automap'])
    .optional()
    .default('continue'),
  columnRemovalStrategy: Yup.string<ColumnRemovalStrategy>()
    .oneOf(['halt', 'continue'])
    .optional()
    .default('continue'),
});
export type PostgresSourceOptionsFormValues = Yup.InferType<
  typeof PostgresSourceOptionsFormValues
>;

const MysqlSourceOptionsFormValues = Yup.object({
  haltOnNewColumnAddition: Yup.boolean().optional().default(false),
  columnRemovalStrategy: Yup.string<ColumnRemovalStrategy>()
    .oneOf(['halt', 'continue'])
    .optional()
    .default('continue'),
});
export type MysqlSourceOptionsFormValues = Yup.InferType<
  typeof MysqlSourceOptionsFormValues
>;

const MssqlSourceOptionsFormValues = Yup.object({
  haltOnNewColumnAddition: Yup.boolean().optional().default(false),
  columnRemovalStrategy: Yup.string<ColumnRemovalStrategy>()
    .oneOf(['halt', 'continue'])
    .optional()
    .default('continue'),
});
export type MssqlSourceOptionsFormValues = Yup.InferType<
  typeof MssqlSourceOptionsFormValues
>;

const DynamoDBSourceUnmappedTransformConfigFormValues = Yup.object({
  byte: JobMappingTransformerForm.required(
    'A default transformer config must be provided for the byte data type'
  ),
  boolean: JobMappingTransformerForm.required(
    'A default transformer config must be provided for the boolean data type'
  ),
  n: JobMappingTransformerForm.required(
    'A default transformer config must be provided for the number data type'
  ),
  s: JobMappingTransformerForm.required(
    'A default transformer config must be provided for the string data type'
  ),
});
export type DynamoDBSourceUnmappedTransformConfigFormValues = Yup.InferType<
  typeof DynamoDBSourceUnmappedTransformConfigFormValues
>;

const DynamoDBSourceOptionsFormValues = Yup.object({
  unmappedTransformConfig:
    DynamoDBSourceUnmappedTransformConfigFormValues.required(
      'Must provide a DynamoDB unmapped transform config'
    ),
  enableConsistentRead: Yup.bool().default(false),
});
export type DynamoDBSourceOptionsFormValues = Yup.InferType<
  typeof DynamoDBSourceOptionsFormValues
>;

export const SourceOptionsFormValues = Yup.object({
  postgres: PostgresSourceOptionsFormValues.optional(),
  mysql: MysqlSourceOptionsFormValues.optional(),
  dynamodb: DynamoDBSourceOptionsFormValues.optional(),
  mssql: MssqlSourceOptionsFormValues.optional(),
});

export type SourceOptionsFormValues = Yup.InferType<
  typeof SourceOptionsFormValues
>;

export const SourceFormValues = Yup.object({
  sourceId: Yup.string().required('Source is required').uuid(),
  // strict().noUnknown() seems to fix an issue with the discriminating types sometimes being seen as present, which results in bad validation.
  sourceOptions: SourceOptionsFormValues.strict()
    .noUnknown()
    .required('Source Options is required'),
});

const DynamoDbDestinationOptionsFormValues = Yup.object({
  tableMappings: Yup.array()
    .of(
      Yup.object({
        sourceTable: Yup.string().required('Source table is required'),
        destinationTable: Yup.string().required(
          'Destination table is required'
        ),
      }).required('The Destinatino form values are required.')
    )
    .required('The Destination form values are required.')
    .default([]),
});
type DynamoDbDestinationOptionsFormValues = Yup.InferType<
  typeof DynamoDbDestinationOptionsFormValues
>;

const PostgresDbDestinationConflictStrategy = Yup.object({
  onConflictDoNothing: Yup.boolean().default(false),
  onConflictDoUpdate: Yup.boolean().default(false),
}).test(
  'only-one-conflict-strategy',
  'Only one conflict strategy can be enabled at a time',
  (value) => {
    if (!value) return true;
    return !(value.onConflictDoNothing && value.onConflictDoUpdate);
  }
);

const PostgresDbDestinationOptionsFormValues = Yup.object({
  truncateBeforeInsert: Yup.boolean().optional().default(false),
  truncateCascade: Yup.boolean().optional().default(false),
  initTableSchema: Yup.boolean().optional().default(false),
  conflictStrategy: PostgresDbDestinationConflictStrategy.optional(),
  skipForeignKeyViolations: Yup.boolean().optional().default(false),
  maxInFlight: Yup.number()
    .min(1, 'Must be greater than or equal to 1')
    .max(200, 'Must be less than or equal to 200') // arbitrarily setting this value here.
    .optional(),
  batch: BatchFormValues.optional(),
});
export type PostgresDbDestinationOptionsFormValues = Yup.InferType<
  typeof PostgresDbDestinationOptionsFormValues
>;

const MysqlDbDestinationConflictStrategy = Yup.object({
  onConflictDoNothing: Yup.boolean().default(false),
  onConflictDoUpdate: Yup.boolean().default(false),
}).test(
  'only-one-conflict-strategy',
  'Only one conflict strategy can be enabled at a time',
  (value) => {
    if (!value) return true;
    return !(value.onConflictDoNothing && value.onConflictDoUpdate);
  }
);

const MysqlDbDestinationOptionsFormValues = Yup.object({
  truncateBeforeInsert: Yup.boolean().optional().default(false),
  initTableSchema: Yup.boolean().optional().default(false),
  conflictStrategy: MysqlDbDestinationConflictStrategy.optional(),
  skipForeignKeyViolations: Yup.boolean().optional().default(false),
  maxInFlight: Yup.number()
    .min(1, 'Must be greater than or equal to 1')
    .max(200, 'Must be less than or equal to 200') // arbitrarily setting this value here.
    .optional(),
  batch: BatchFormValues.optional(),
});
export type MysqlDbDestinationOptionsFormValues = Yup.InferType<
  typeof MysqlDbDestinationOptionsFormValues
>;

const MssqlDbDestinationOptionsFormValues = Yup.object({
  truncateBeforeInsert: Yup.boolean().optional().default(false),
  initTableSchema: Yup.boolean().optional().default(false),
  onConflictDoNothing: Yup.boolean().optional().default(false),
  skipForeignKeyViolations: Yup.boolean().optional().default(false),
  maxInFlight: Yup.number()
    .min(1, 'Must be greater than or equal to 1')
    .max(200, 'Must be less than or equal to 200') // arbitrarily setting this value here.
    .optional(),
  batch: BatchFormValues.optional(),
});
export type MssqlDbDestinationOptionsFormValues = Yup.InferType<
  typeof MssqlDbDestinationOptionsFormValues
>;

export const AwsS3DestinationOptionsFormValues = Yup.object({
  storageClass: Yup.number().optional(),
  maxInFlight: Yup.number()
    .min(1, 'Must be greater than or equal to 1')
    .max(200, 'Must be less than or equal to 200') // arbitrarily setting this value here.
    .optional(),
  timeout: Yup.string()
    .optional()
    .test('duration', 'Must be a valid duration', getDurationValidateFn()),
  batch: BatchFormValues.optional(),
});
export type AwsS3DestinationOptionsFormValues = Yup.InferType<
  typeof AwsS3DestinationOptionsFormValues
>;

export const DestinationOptionsFormValues = Yup.object({
  postgres: PostgresDbDestinationOptionsFormValues.optional(),
  mysql: MysqlDbDestinationOptionsFormValues.optional(),
  dynamodb: DynamoDbDestinationOptionsFormValues.optional(),
  mssql: MssqlDbDestinationOptionsFormValues.optional(),
  awss3: AwsS3DestinationOptionsFormValues.optional(),
}).required('Destination Options are required.');
// Object that holds connection specific destination options for a job
export type DestinationOptionsFormValues = Yup.InferType<
  typeof DestinationOptionsFormValues
>;

const SchemaFormValuesDestinationOptions = Yup.object({
  destinationId: Yup.string().required('Destination Connection is required.'), // in this case it is the connection id
  dynamodb: DynamoDbDestinationOptionsFormValues.optional(),
});
export type SchemaFormValuesDestinationOptions = Yup.InferType<
  typeof SchemaFormValuesDestinationOptions
>;

export const SchemaFormValues = Yup.object({
  mappings: Yup.array()
    .of(JobMappingFormValues)
    .required('Schema mappings are required.'),
  virtualForeignKeys: Yup.array().of(VirtualForeignConstraintFormValues),
  connectionId: Yup.string().required('Connection is required.'),

  destinationOptions: Yup.array()
    .of(
      SchemaFormValuesDestinationOptions.required(
        'Schems Form Destination options are required'
      )
    )
    .required('Destination options are required'),
});
export type SchemaFormValues = Yup.InferType<typeof SchemaFormValues>;

export const NewDestinationFormValues = Yup.object({
  connectionId: Yup.string()
    .required('Connection is required')
    .uuid()
    .test(
      'checkConnectionUnique',
      'Destination must be different from source.',
      (value, ctx) => {
        if (ctx.from) {
          const { sourceId } = ctx.from[ctx.from.length - 1].value;
          if (value === sourceId) {
            return false;
          }
        }

        return true;
      }
    ),
  destinationOptions: DestinationOptionsFormValues,
}).required('Destination form values are required.');
export type NewDestinationFormValues = Yup.InferType<
  typeof NewDestinationFormValues
>;

const EditDestinationOptionsFormValues = Yup.object({
  destinationId: Yup.string().required('Destination is required'),
})
  .concat(
    DestinationOptionsFormValues.required(
      'Destination option form values are required.'
    )
  )
  .required('Destination and destination option form values are required.');
export type EditDestinationOptionsFormValues = Yup.InferType<
  typeof EditDestinationOptionsFormValues
>;

export const DataSyncSourceFormValues = SourceFormValues.concat(
  SchemaFormValues
).concat(
  Yup.object({
    destinationOptions: Yup.array()
      .of(
        EditDestinationOptionsFormValues.required(
          'Destination option form values are required.'
        )
      )
      .required('Destination options are required'),
  })
);
export type DataSyncSourceFormValues = Yup.InferType<
  typeof DataSyncSourceFormValues
>;

export const DefaultTransformerFormValues = Yup.object({
  overrideTransformers: Yup.boolean().default(false),
});

export type DefaultTransformerFormValues = Yup.InferType<
  typeof DefaultTransformerFormValues
>;

export function toJobSourcePostgresNewColumnAdditionStrategy(
  strategy?: NewColumnAdditionStrategy
): PostgresSourceConnectionOptions_NewColumnAdditionStrategy | undefined {
  switch (strategy) {
    case 'continue': {
      return undefined;
    }
    case 'automap': {
      return create(
        PostgresSourceConnectionOptions_NewColumnAdditionStrategySchema,
        {
          strategy: {
            case: 'autoMap',
            value: create(
              PostgresSourceConnectionOptions_NewColumnAdditionStrategy_AutoMapSchema
            ),
          },
        }
      );
    }
    case 'halt': {
      return create(
        PostgresSourceConnectionOptions_NewColumnAdditionStrategySchema,
        {
          strategy: {
            case: 'haltJob',
            value: create(
              PostgresSourceConnectionOptions_NewColumnAdditionStrategy_HaltJobSchema
            ),
          },
        }
      );
    }
    default: {
      return undefined;
    }
  }
}
export function toNewColumnAdditionStrategy(
  input: PostgresSourceConnectionOptions_NewColumnAdditionStrategy | undefined
): NewColumnAdditionStrategy {
  switch (input?.strategy.case) {
    case 'haltJob': {
      return 'halt';
    }
    case 'autoMap': {
      return 'automap';
    }
    default: {
      return 'continue';
    }
  }
}

export function toJobSourcePostgresColumnRemovalStrategy(
  strategy?: ColumnRemovalStrategy
): PostgresSourceConnectionOptions_ColumnRemovalStrategy | undefined {
  switch (strategy) {
    case 'continue': {
      return undefined;
    }
    case 'halt': {
      return create(
        PostgresSourceConnectionOptions_ColumnRemovalStrategySchema,
        {
          strategy: {
            case: 'haltJob',
            value: create(
              PostgresSourceConnectionOptions_ColumnRemovalStrategy_HaltJobSchema
            ),
          },
        }
      );
    }
    default: {
      return undefined;
    }
  }
}

export function toJobSourceMysqlColumnRemovalStrategy(
  strategy?: ColumnRemovalStrategy
): MysqlSourceConnectionOptions_ColumnRemovalStrategy | undefined {
  switch (strategy) {
    case 'continue': {
      return undefined;
    }
    case 'halt': {
      return create(MysqlSourceConnectionOptions_ColumnRemovalStrategySchema, {
        strategy: {
          case: 'haltJob',
          value: create(
            MysqlSourceConnectionOptions_ColumnRemovalStrategy_HaltJobSchema
          ),
        },
      });
    }
    default: {
      return undefined;
    }
  }
}

export function toJobSourceMssqlColumnRemovalStrategy(
  strategy?: ColumnRemovalStrategy
): MssqlSourceConnectionOptions_ColumnRemovalStrategy | undefined {
  switch (strategy) {
    case 'continue': {
      return undefined;
    }
    case 'halt': {
      return create(MssqlSourceConnectionOptions_ColumnRemovalStrategySchema, {
        strategy: {
          case: 'haltJob',
          value: create(
            MssqlSourceConnectionOptions_ColumnRemovalStrategy_HaltJobSchema
          ),
        },
      });
    }
    default: {
      return undefined;
    }
  }
}

export function toColumnRemovalStrategy(
  input:
    | PostgresSourceConnectionOptions_ColumnRemovalStrategy
    | MysqlSourceConnectionOptions_ColumnRemovalStrategy
    | MssqlSourceConnectionOptions_ColumnRemovalStrategy
    | undefined
): ColumnRemovalStrategy {
  switch (input?.strategy.case) {
    case 'haltJob': {
      return 'halt';
    }
    default: {
      return 'continue';
    }
  }
}
