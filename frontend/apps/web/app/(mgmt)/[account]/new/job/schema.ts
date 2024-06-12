import { RESOURCE_NAME_REGEX } from '@/yup-validations/connections';
import {
  DestinationFormValues,
  JobMappingFormValues,
  SchemaFormValues,
  SourceFormValues,
} from '@/yup-validations/jobs';
import { Connection, IsJobNameAvailableResponse } from '@neosync/sdk';
import cron from 'cron-validate';
import * as Yup from 'yup';

export type NewJobType = 'data-sync' | 'generate-table' | 'ai-generate-table';

// Schema for a job's workflow settings
export const WorkflowSettingsSchema = Yup.object({
  runTimeout: Yup.number().optional().min(0),
});

export type WorkflowSettingsSchema = Yup.InferType<
  typeof WorkflowSettingsSchema
>;

export const ActivityOptionsSchema = Yup.object({
  scheduleToCloseTimeout: Yup.number()
    .optional()
    .min(0)
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
    .min(0)
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
    maximumAttempts: Yup.number().optional().min(0),
  }).optional(),
});

export type ActivityOptionsSchema = Yup.InferType<typeof ActivityOptionsSchema>;

export const DefineFormValues = Yup.object({
  jobName: Yup.string()
    .trim()
    .required('Name is a required field')
    .min(3)
    .max(30)
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
        const res = await isJobNameAvailable(value, accountId);
        return res.isAvailable;
      }
    ),
  cronSchedule: Yup.string()
    .optional()
    .test('validateCron', 'Not a valid cron schedule', (value, context) => {
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
    }),
  initiateJobRun: Yup.boolean(),
  workflowSettings: WorkflowSettingsSchema.optional(),
  syncActivityOptions: ActivityOptionsSchema.optional(),
});

export type DefineFormValues = Yup.InferType<typeof DefineFormValues>;

export const ConnectFormValues = SourceFormValues.concat(
  Yup.object({
    destinations: Yup.array(DestinationFormValues).required(),
  })
).test(
  'unique-connections',
  'connections must be unique and type specific', // this message isn't exposed anywhere
  function (value, ctx) {
    const connections: Connection[] = ctx.options.context?.connections ?? [];

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

    if (
      destinationIds.some(
        (destId) => !isValidConnectionPair(value.sourceId, destId, connections)
      )
    ) {
      destinationIds.forEach((destId, idx) => {
        if (!isValidConnectionPair(value.sourceId, destId, connections)) {
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
        }
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

function isValidConnectionPair(
  connId1: string,
  connId2: string,
  connections: Connection[]
): boolean {
  const conn1 = connections.find((c) => c.id === connId1);
  const conn2 = connections.find((c) => c.id === connId2);

  if (!conn1 || !conn2) {
    return true;
  }
  if (
    conn1.connectionConfig?.config.case === 'awsS3Config' ||
    conn2.connectionConfig?.config.case === 'awsS3Config'
  ) {
    return true;
  }

  if (
    conn1.connectionConfig?.config.case === conn2.connectionConfig?.config.case
  ) {
    return true;
  }

  return false;
}

function getErrorConnectionTypes(
  isSource: boolean,
  connId: string,
  connections: Connection[]
): string {
  const conn = connections.find((c) => c.id === connId);
  if (!conn) {
    return isSource ? '[Postgres, Mysql]' : '[Postgres, Mysql, AWS S3]';
  }
  if (conn.connectionConfig?.config.case === 'awsS3Config') {
    return '[Mysql, Postgres]';
  }
  if (conn.connectionConfig?.config.case === 'mysqlConfig') {
    return isSource ? '[Postgres]' : '[Mysql, AWS S3]';
  }

  if (conn.connectionConfig?.config.case === 'pgConfig') {
    return isSource ? '[Mysql]' : '[Postgres, AWS S3]';
  }
  return '';
}

const SINGLE_SUBSET_FORM_SCHEMA = Yup.object({
  schema: Yup.string().trim().required(),
  table: Yup.string().trim().required(),
  whereClause: Yup.string().trim().optional(),
});

export const SingleTableConnectFormValues = Yup.object({
  fkSourceConnectionId: Yup.string().required('Connection is required').uuid(),
  destination: DestinationFormValues,
});
export type SingleTableConnectFormValues = Yup.InferType<
  typeof SingleTableConnectFormValues
>;

export const SingleTableAiConnectFormValues = Yup.object({
  sourceId: Yup.string().required('Connection is required').uuid(),
  fkSourceConnectionId: Yup.string().required('Connection is required').uuid(),
  destination: DestinationFormValues,
});

export type SingleTableAiConnectFormValues = Yup.InferType<
  typeof SingleTableAiConnectFormValues
>;

export const SingleTableAiSchemaFormValues = Yup.object({
  numRows: Yup.number()
    .required('Must provide a number of rows to generate')
    .min(1)
    .max(1000)
    .default(10),
  model: Yup.string().required('must provide a model name to use.'),
  userPrompt: Yup.string(),

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
  }).required(),

  schema: Yup.object({
    numRows: Yup.number()
      .required('Must provide a number of rows to generate')
      .min(1)
      .max(1000)
      .default(10),
    model: Yup.string().required('must provide a model name to use.'),
    userPrompt: Yup.string(),

    schema: Yup.string().required('Must provide a valid schema'),
    table: Yup.string().required('Must provide a valid table'),
  }).required(),
});
export type SingleTableEditAiSourceFormValues = Yup.InferType<
  typeof SingleTableEditAiSourceFormValues
>;

export const SingleTableSchemaFormValues = Yup.object({
  numRows: Yup.number().required().min(1),
  mappings: Yup.array().of(JobMappingFormValues).required(),
});
export type SingleTableSchemaFormValues = Yup.InferType<
  typeof SingleTableSchemaFormValues
>;

export const SingleTableEditSourceFormValues = Yup.object({
  source: Yup.object({
    fkSourceConnectionId: Yup.string()
      .required('Connection is required')
      .uuid(),
  }).required(),

  numRows: Yup.number()
    .required('Must provide a number of rows to generate')
    .min(1)
    .default(10),
  mappings: Yup.array().of(JobMappingFormValues).required(),
}).required();

export type SingleTableEditSourceFormValues = Yup.InferType<
  typeof SingleTableEditSourceFormValues
>;

export const SubsetFormValues = Yup.object({
  subsets: Yup.array(SINGLE_SUBSET_FORM_SCHEMA).required(),
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
}).required();
export type CreateJobFormValues = Yup.InferType<typeof CreateJobFormValues>;

export const CreateSingleTableGenerateJobFormValues = Yup.object({
  define: DefineFormValues,
  connect: SingleTableConnectFormValues,
  schema: SingleTableSchemaFormValues,
}).required();
export type CreateSingleTableGenerateJobFormValues = Yup.InferType<
  typeof CreateSingleTableGenerateJobFormValues
>;

async function isJobNameAvailable(
  name: string,
  accountId: string
): Promise<IsJobNameAvailableResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/is-job-name-available?name=${name}`,
    {
      method: 'GET',
      headers: {
        'content-type': 'application/json',
      },
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return IsJobNameAvailableResponse.fromJson(await res.json());
}
