import { convertMinutesToNanoseconds } from '@/util/util';
import {
  convertJobMappingTransformerFormToJobMappingTransformer,
  toJobDestinationOptions,
  toMysqlSourceSchemaOptions,
  toPostgresSourceSchemaOptions,
} from '@/yup-validations/jobs';
import {
  ActivityOptions,
  Connection,
  CreateJobRequest,
  CreateJobResponse,
  JobDestination,
  JobMapping,
  JobSource,
  JobSourceOptions,
  MongoDBSourceConnectionOptions,
  MysqlSourceConnectionOptions,
  PostgresSourceConnectionOptions,
  RetryPolicy,
  WorkflowOptions,
} from '@neosync/sdk';
import { FormValues } from '../schema';

export async function createNewJob(
  formData: FormValues,
  accountId: string,
  connections: Connection[]
): Promise<CreateJobResponse> {
  const connectionIdMap = new Map(
    connections.map((connection) => [connection.id, connection])
  );
  const sourceConnection = connections.find(
    (c) => c.id === formData.connect.sourceId
  );

  let workflowOptions: WorkflowOptions | undefined = undefined;
  if (formData.define.workflowSettings?.runTimeout) {
    workflowOptions = new WorkflowOptions({
      runTimeout: convertMinutesToNanoseconds(
        formData.define.workflowSettings.runTimeout
      ),
    });
  }
  let syncOptions: ActivityOptions | undefined = undefined;
  if (formData.define.syncActivityOptions) {
    const formSyncOpts = formData.define.syncActivityOptions;
    syncOptions = new ActivityOptions({
      scheduleToCloseTimeout:
        formSyncOpts.scheduleToCloseTimeout !== undefined
          ? convertMinutesToNanoseconds(formSyncOpts.scheduleToCloseTimeout)
          : undefined,
      startToCloseTimeout:
        formSyncOpts.startToCloseTimeout !== undefined
          ? convertMinutesToNanoseconds(formSyncOpts.startToCloseTimeout)
          : undefined,
      retryPolicy: new RetryPolicy({
        maximumAttempts: formSyncOpts.retryPolicy?.maximumAttempts,
      }),
    });
  }

  const body = new CreateJobRequest({
    accountId,
    jobName: formData.define.jobName,
    cronSchedule: formData.define.cronSchedule,
    initiateJobRun: formData.define.initiateJobRun,
    mappings: formData.schema.mappings.map((m) => {
      return new JobMapping({
        schema: m.schema,
        table: m.table,
        column: m.column,
        transformer: convertJobMappingTransformerFormToJobMappingTransformer(
          m.transformer
        ),
      });
    }),
    source: new JobSource({
      options: toJobSourceOptions(formData, sourceConnection),
    }),
    destinations: formData.connect.destinations.map((d) => {
      return new JobDestination({
        connectionId: d.connectionId,
        options: toJobDestinationOptions(
          d,
          connectionIdMap.get(d.connectionId)
        ),
      });
    }),
    workflowOptions: workflowOptions,
    syncOptions: syncOptions,
  });

  function toJobSourceOptions(
    values: FormValues,
    connection?: Connection
  ): JobSourceOptions {
    if (!connection) {
      return new JobSourceOptions();
    }
    switch (connection.connectionConfig?.config.case) {
      case 'pgConfig':
        return new JobSourceOptions({
          config: {
            case: 'postgres',
            value: new PostgresSourceConnectionOptions({
              connectionId: formData.connect.sourceId,
              haltOnNewColumnAddition:
                values.connect.sourceOptions.haltOnNewColumnAddition,
              subsetByForeignKeyConstraints:
                values.subset?.subsetOptions.subsetByForeignKeyConstraints,
              schemas:
                values.subset?.subsets &&
                toPostgresSourceSchemaOptions(values.subset.subsets),
            }),
          },
        });
      case 'mysqlConfig':
        return new JobSourceOptions({
          config: {
            case: 'mysql',
            value: new MysqlSourceConnectionOptions({
              connectionId: formData.connect.sourceId,
              haltOnNewColumnAddition:
                values.connect.sourceOptions.haltOnNewColumnAddition,
              subsetByForeignKeyConstraints:
                values.subset?.subsetOptions.subsetByForeignKeyConstraints,
              schemas:
                values.subset?.subsets &&
                toMysqlSourceSchemaOptions(values.subset?.subsets),
            }),
          },
        });
      case 'mongoConfig':
        return new JobSourceOptions({
          config: {
            case: 'mongodb',
            value: new MongoDBSourceConnectionOptions({
              connectionId: formData.connect.sourceId,
            }),
          },
        });
      default:
        throw new Error('unsupported connection type');
    }
  }

  const res = await fetch(`/api/accounts/${accountId}/jobs`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }

  return CreateJobResponse.fromJson(await res.json());
}
