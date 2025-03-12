import { RESOURCE_NAME_REGEX } from '@/yup-validations/connections';
import {
  JobMappingFormValues,
  NewDestinationFormValues,
  SchemaFormValues,
  SourceFormValues,
} from '@/yup-validations/jobs';
import { create, MessageInitShape } from '@bufbuild/protobuf';
import {
  ConnectError,
  Connection,
  IsJobNameAvailableRequestSchema,
  IsJobNameAvailableResponse,
} from '@neosync/sdk';
import { UseMutateAsyncFunction } from '@tanstack/react-query';
import cron from 'cron-validate';
import * as Yup from 'yup';
import { isValidConnectionPair } from '../../connections/util';

export type NewJobType =
  | 'data-sync'
  | 'generate-table'
  | 'ai-generate-table'
  | 'pii-detection';

// Schema for a job's workflow settings
export const WorkflowSettingsSchema = Yup.object({
  runTimeout: Yup.number()
    .optional()
    .min(0, 'The Job Run Timeout cannot be less than 0 minutes'),
});

export type WorkflowSettingsSchema = Yup.InferType<
  typeof WorkflowSettingsSchema
>;

export const ActivityOptionsFormValues = Yup.object({
  scheduleToCloseTimeout: Yup.number()
    .optional()
    .min(
      0,
      'The Max Table Timeout (including Retries) cannot be less than 0 minutes'
    )
    .test(
      'non-zero-both',
      'Both timeouts cannot be 0',
      function (value, context) {
        // Checking the other field's value from the context
        const startToCloseTimeout = context.parent.startToCloseTimeout;
        return !(value === 0 && startToCloseTimeout === 0);
      }
    ),
  startToCloseTimeout: Yup.number()
    .optional()
    .min(0, 'The Table Sync cannot be less than 0 minutes')
    .test(
      'non-zero-both',
      'Both timeouts cannot be 0',
      function (value, context) {
        // Checking the other field's value from the context
        const scheduleToCloseTimeout = context.parent.scheduleToCloseTimeout;
        return !(value === 0 && scheduleToCloseTimeout === 0);
      }
    ),
  retryPolicy: Yup.object({
    maximumAttempts: Yup.number()
      .optional()
      .min(0, 'The Maximum Retry Attempts cannot be less than 0'),
  }).optional(),
});

export type ActivityOptionsFormValues = Yup.InferType<
  typeof ActivityOptionsFormValues
>;

export const DefineFormValues = Yup.object({
  jobName: Yup.string()
    .trim()
    .required('Name is a required field')
    .min(3, 'The Job Name must be at least 3 characters')
    .max(100, 'The Job name must be less than or equal to 100 characters')
    .test(
      'checkNameUnique',
      'This name is already taken.',
      async (value, context) => {
        if (!value || value.length < 3) {
          return false;
        }
        if (!RESOURCE_NAME_REGEX.test(value)) {
          return context.createError({
            message:
              'Job Name can only include lowercase letters, numbers, and hyphens',
          });
        }
        const accountId = context.options.context?.accountId;
        if (!accountId) {
          return false;
        }
        const isJobNameAvailable:
          | UseMutateAsyncFunction<
              IsJobNameAvailableResponse,
              ConnectError,
              MessageInitShape<typeof IsJobNameAvailableRequestSchema>,
              unknown
            >
          | undefined = context?.options?.context?.isJobNameAvailable;
        if (isJobNameAvailable) {
          const res = await isJobNameAvailable(
            create(IsJobNameAvailableRequestSchema, {
              accountId,
              name: value,
            })
          );
          if (!res.isAvailable) {
            return context.createError({
              message: 'This Job Name is already taken.',
            });
          }
        }
        return true;
      }
    ),
  cronSchedule: Yup.string()
    .optional()
    .test(
      'validateCron',
      'The Schedule must be a valid Cron string',
      (value, context) => {
        if (!value) {
          return true;
        }
        const output = cron(value);
        if (output.isValid()) {
          return true;
        }
        if (output.isError()) {
          const errors = output.getError();
          if (errors.length > 0) {
            return context.createError({ message: errors.join(', ') });
          }
        }
        return output.isValid();
      }
    ),
  initiateJobRun: Yup.boolean(),
  workflowSettings: WorkflowSettingsSchema.optional(),
  syncActivityOptions: ActivityOptionsFormValues.optional(),
});

export type DefineFormValues = Yup.InferType<typeof DefineFormValues>;

export const ConnectFormValues = SourceFormValues.concat(
  Yup.object({
    destinations: Yup.array(NewDestinationFormValues).required(
      'At least one destination Connection is required'
    ),
  })
).test(
  // todo: need to add a test for generate / ai generate too
  'unique-connections',
  'Connections must be unique and the same Connection type', // this message isn't exposed anywhere
  function (value, ctx) {
    const connections: Connection[] = ctx.options.context?.connections ?? [];
    const connectionsRecord = connections.reduce(
      (record, conn) => {
        record[conn.id] = conn;
        return record;
      },
      {} as Record<string, Connection>
    );

    const destinationIds = value.destinations.map((dst) => dst.connectionId);

    const errors: Yup.ValidationError[] = [];

    if (destinationIds.some((destId) => value.sourceId === destId)) {
      errors.push(
        ctx.createError({
          path: 'sourceId',
          message: 'Source must be different from destination',
        })
      );
    }

    const sourceConn = connectionsRecord[value.sourceId];
    if (!sourceConn) {
      errors.push(
        ctx.createError({
          path: 'sourceId',
          message: 'Source is not a valid connection',
        })
      );
      return new Yup.ValidationError(errors);
    }

    const invalidDestinationConnections = destinationIds
      .map((destId) => connectionsRecord[destId])
      .filter((dest) => !!dest && !isValidConnectionPair(sourceConn, dest));
    if (invalidDestinationConnections.length > 0) {
      const invalidDestRecord = invalidDestinationConnections.reduce(
        (record, dest) => {
          record[dest.id] = dest;
          return record;
        },
        {} as Record<string, Connection>
      );
      destinationIds.forEach((destId, idx) => {
        const invalidDest = invalidDestRecord[destId];
        if (!invalidDest) {
          return;
        }
        errors.push(
          ctx.createError({
            path: `destinations.${idx}.connectionId`,
            message: `Destination connection type must be one of: ${getErrorConnectionTypes(
              false,
              value.sourceId,
              connections
            )}`,
          })
        );
      });
    }

    if (destinationIds.length !== new Set(destinationIds).size) {
      const valueIdxMap = new Map<string, number[]>();
      destinationIds.forEach((dstId, idx) => {
        const idxs = valueIdxMap.get(dstId);
        if (idxs !== undefined) {
          idxs.push(idx);
        } else {
          valueIdxMap.set(dstId, [idx]);
        }
      });

      Array.from(valueIdxMap.values()).forEach((indices) => {
        if (indices.length > 1) {
          indices.forEach((idx) =>
            errors.push(
              ctx.createError({
                path: `destinations.${idx}.connectionId`,
                message:
                  'Destination connections must be unique from one another',
              })
            )
          );
        }
      });
    }

    if (errors.length > 0) {
      return new Yup.ValidationError(errors);
    }
    return true;
  }
);
export type ConnectFormValues = Yup.InferType<typeof ConnectFormValues>;

// todo: move this and centralize / automate using new maps
function getErrorConnectionTypes(
  isSource: boolean,
  connId: string,
  connections: Connection[]
): string {
  const conn = connections.find((c) => c.id === connId);
  if (!conn) {
    return isSource
      ? '[Postgres, Mysql, MongoDB, DynamoDB]'
      : '[Postgres, Mysql, MongoDB, DynamoDB, AWS S3, GCP Cloud Storage]';
  }
  if (
    conn.connectionConfig?.config.case === 'awsS3Config' ||
    conn.connectionConfig?.config.case === 'gcpCloudstorageConfig'
  ) {
    return '[Mysql, Postgres]';
  }
  if (conn.connectionConfig?.config.case === 'mysqlConfig') {
    return isSource ? '[Postgres]' : '[Mysql, AWS S3, GCP Cloud Storage]';
  }

  if (conn.connectionConfig?.config.case === 'pgConfig') {
    return isSource ? '[Mysql]' : '[Postgres, AWS S3, GCP Cloud Storage]';
  }
  if (conn.connectionConfig?.config.case === 'dynamodbConfig') {
    return isSource ? '[DynamoDB]' : '[DynamoDB]';
  }
  return '';
}

const SingleSubsetFormValue = Yup.object({
  schema: Yup.string().trim().required('A schema is required'),
  table: Yup.string().trim().required('A table is required'),
  whereClause: Yup.string().trim().optional(),
});
export type SingleSubsetFormValue = Yup.InferType<typeof SingleSubsetFormValue>;

export const SingleTableConnectFormValues = Yup.object({
  fkSourceConnectionId: Yup.string().required('Connection is required').uuid(),
  destination: NewDestinationFormValues,
});
export type SingleTableConnectFormValues = Yup.InferType<
  typeof SingleTableConnectFormValues
>;

export const SingleTableAiConnectFormValues = Yup.object({
  sourceId: Yup.string().required('Connection is required').uuid(),
  fkSourceConnectionId: Yup.string().required('Connection is required').uuid(),
  destination: NewDestinationFormValues,
});

export type SingleTableAiConnectFormValues = Yup.InferType<
  typeof SingleTableAiConnectFormValues
>;

export const PiiDetectionConnectFormValues = Yup.object().shape({
  sourceId: Yup.string().required('Connection is required').uuid(),
});
export type PiiDetectionConnectFormValues = Yup.InferType<
  typeof PiiDetectionConnectFormValues
>;

const TableScanFilterModeFormValue = Yup.string()
  .required()
  .oneOf(['include_all', 'include', 'exclude']);
export type TableScanFilterModeFormValue = Yup.InferType<
  typeof TableScanFilterModeFormValue
>;

const FilterPatternTableIdentifier = Yup.object().shape({
  schema: Yup.string().required(),
  table: Yup.string().required(),
});
export type FilterPatternTableIdentifier = Yup.InferType<
  typeof FilterPatternTableIdentifier
>;

const TableScanFilterPatternsFormValue = Yup.object().shape({
  schemas: Yup.array().of(Yup.string().required()).required().default([]),
  tables: Yup.array()
    .of(FilterPatternTableIdentifier.required())
    .required()
    .default([]),
});
export type TableScanFilterPatternsFormValue = Yup.InferType<
  typeof TableScanFilterPatternsFormValue
>;

const TableScanFilterFormValue = Yup.object().shape({
  mode: TableScanFilterModeFormValue,
  patterns: TableScanFilterPatternsFormValue,
});
export type TableScanFilterFormValue = Yup.InferType<
  typeof TableScanFilterFormValue
>;

export const PiiDetectionSchemaFormValues = Yup.object().shape({
  dataSampling: Yup.object().shape({
    isEnabled: Yup.boolean().required().default(true),
  }),
  tableScanFilter: TableScanFilterFormValue,
  userPrompt: Yup.string(),
});

export type PiiDetectionSchemaFormValues = Yup.InferType<
  typeof PiiDetectionSchemaFormValues
>;

export const SingleTableAiSchemaFormValues = Yup.object({
  numRows: Yup.number()
    .required('Must provide a number of rows to generate')
    .min(1, 'Must be at least 1')
    .max(1000, 'The number of rows must be less than or equal to 1000')
    .default(10),
  model: Yup.string().required('must provide a model name to use.'),
  userPrompt: Yup.string(),
  generateBatchSize: Yup.number()
    .required('Must provide a batch size when generating rows')
    .min(1, 'Must be at least 1')
    .max(100, 'The Batch Size must be less than or equal to 100.')
    .default(10)
    .test(
      'batch-size-num-rows',
      'batch size must always be smaller than the number of rows',
      function (value, context) {
        return value <= context.parent.numRows;
      }
    ),

  schema: Yup.string().required('Must provide a valid schema'),
  table: Yup.string().required('Must provide a valid table'),
});

export type SingleTableAiSchemaFormValues = Yup.InferType<
  typeof SingleTableAiSchemaFormValues
>;

export const SingleTableEditAiSourceFormValues = Yup.object({
  source: Yup.object({
    sourceId: Yup.string().required('Connection is required').uuid(),
    fkSourceConnectionId: Yup.string()
      .required('Connection is required')
      .uuid(),
  }).required('AI Generate form values are required.'),

  schema: Yup.object({
    numRows: Yup.number()
      .required('Must provide a number of rows to generate')
      .min(1, 'Must be at least 1')
      .max(1000, 'The number of rows must be less than or equal to 1000.')
      .default(10),
    model: Yup.string().required('must provide a model name to use.'),
    userPrompt: Yup.string(),
    generateBatchSize: Yup.number()
      .required('Must provide a batch size when generating rows')
      .min(1, 'Must be at least 1')
      .max(100, 'The Batch Size must be less than or equal to 100.')
      .default(10)
      .test(
        'batch-size-num-rows',
        'batch size must always be smaller than the number of rows',
        function (value, context) {
          return value <= context.parent.numRows;
        }
      ),

    schema: Yup.string().required('Must provide a valid schema'),
    table: Yup.string().required('Must provide a valid table'),
  }).required('AI Generate schema values are required.'),
});
export type SingleTableEditAiSourceFormValues = Yup.InferType<
  typeof SingleTableEditAiSourceFormValues
>;

export const SingleTableSchemaFormValues = Yup.object({
  numRows: Yup.number()
    .required('THe number of rows to generate is required')
    .min(
      1,
      'The number of rows to generate must be greater than or equal to 1.'
    ),
  mappings: Yup.array()
    .of(JobMappingFormValues)
    .required('Table Mappings are required.'),
});
export type SingleTableSchemaFormValues = Yup.InferType<
  typeof SingleTableSchemaFormValues
>;

export const SingleTableEditSourceFormValues = Yup.object({
  source: Yup.object({
    fkSourceConnectionId: Yup.string()
      .required('Connection is required')
      .uuid(),
  }).required('Source Connection is required.'),
  numRows: Yup.number()
    .required('Must provide a number of rows to generate')
    .min(
      1,
      'The number of rows to generate must be greater than or equal to 1.'
    )
    .default(10),
  mappings: Yup.array()
    .of(JobMappingFormValues)
    .required('Mappings are required.'),
}).required('Generate table form values are required.');

export type SingleTableEditSourceFormValues = Yup.InferType<
  typeof SingleTableEditSourceFormValues
>;

export const SubsetFormValues = Yup.object({
  subsets: Yup.array(SingleSubsetFormValue).required('Subset is required.'),
  subsetOptions: Yup.object({
    subsetByForeignKeyConstraints: Yup.boolean().default(true),
  }),
});

export type SubsetFormValues = Yup.InferType<typeof SubsetFormValues>;

const CreateJobFormValues = Yup.object({
  define: DefineFormValues,
  connect: ConnectFormValues,
  schema: SchemaFormValues,
  subset: SubsetFormValues.optional(),
}).required('Job form values are required.');
export type CreateJobFormValues = Yup.InferType<typeof CreateJobFormValues>;

export const CreateSingleTableGenerateJobFormValues = Yup.object({
  define: DefineFormValues,
  connect: SingleTableConnectFormValues,
  schema: SingleTableSchemaFormValues,
}).required('Generate form values are required.');
export type CreateSingleTableGenerateJobFormValues = Yup.InferType<
  typeof CreateSingleTableGenerateJobFormValues
>;

export const CreateSingleTableAiGenerateJobFormValues = Yup.object({
  define: DefineFormValues,
  connect: SingleTableAiConnectFormValues,
  schema: SingleTableAiSchemaFormValues,
}).required('AI Generate form values are required.');
export type CreateSingleTableAiGenerateJobFormValues = Yup.InferType<
  typeof CreateSingleTableAiGenerateJobFormValues
>;

export const CreatePiiDetectionJobFormValues = Yup.object()
  .shape({
    define: DefineFormValues,
    connect: PiiDetectionConnectFormValues,
    schema: PiiDetectionSchemaFormValues,
  })
  .required('PII Detection form values are required.');
export type CreatePiiDetectionJobFormValues = Yup.InferType<
  typeof CreatePiiDetectionJobFormValues
>;
export interface DefineFormValuesContext {
  accountId: string;
  isJobNameAvailable: UseMutateAsyncFunction<
    IsJobNameAvailableResponse,
    ConnectError,
    MessageInitShape<typeof IsJobNameAvailableRequestSchema>,
    unknown
  >;
}
