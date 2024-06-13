import { convertMinutesToNanoseconds } from '@/util/util';
import {
  DestinationFormValues,
  convertJobMappingTransformerFormToJobMappingTransformer,
} from '@/yup-validations/jobs';
import {
  ActivityOptions,
  AiGenerateSourceOptions,
  AiGenerateSourceSchemaOption,
  AiGenerateSourceTableOption,
  AwsS3DestinationConnectionOptions,
  Connection,
  CreateJobDestinationConnectionsRequest,
  CreateJobDestinationConnectionsResponse,
  CreateJobRequest,
  CreateJobResponse,
  DatabaseTable,
  GenerateSourceOptions,
  GenerateSourceSchemaOption,
  GenerateSourceTableOption,
  GetAiGeneratedDataRequest,
  IsJobNameAvailableResponse,
  JobDestination,
  JobDestinationOptions,
  JobMapping,
  JobSource,
  JobSourceOptions,
  JobSourceSqlSubetSchemas,
  MongoDBDestinationConnectionOptions,
  MongoDBSourceConnectionOptions,
  MysqlDestinationConnectionOptions,
  MysqlOnConflictConfig,
  MysqlSourceConnectionOptions,
  MysqlSourceSchemaOption,
  MysqlSourceSchemaSubset,
  MysqlSourceTableOption,
  MysqlTruncateTableConfig,
  PauseJobRequest,
  PauseJobResponse,
  PostgresDestinationConnectionOptions,
  PostgresOnConflictConfig,
  PostgresSourceConnectionOptions,
  PostgresSourceSchemaOption,
  PostgresSourceSchemaSubset,
  PostgresSourceTableOption,
  PostgresTruncateTableConfig,
  RetryPolicy,
  SetJobSourceSqlConnectionSubsetsRequest,
  SetJobSourceSqlConnectionSubsetsResponse,
  SetJobSyncOptionsRequest,
  SetJobWorkflowOptionsRequest,
  SetJobWorkflowOptionsResponse,
  UpdateJobDestinationConnectionRequest,
  UpdateJobDestinationConnectionResponse,
  UpdateJobScheduleRequest,
  UpdateJobScheduleResponse,
  WorkflowOptions,
} from '@neosync/sdk';
import { SampleRecord } from '../new/job/aigenerate/single/schema/types';
import {
  ActivityOptionsSchema,
  CreateJobFormValues,
  CreateSingleTableAiGenerateJobFormValues,
  CreateSingleTableGenerateJobFormValues,
  SubsetFormValues,
  WorkflowSettingsSchema,
} from '../new/job/schema';

type GetConnectionById = (id: string) => Connection | undefined;

export async function createNewSingleTableAiGenerateJob(
  values: CreateSingleTableAiGenerateJobFormValues,
  accountId: string,
  getConnectionById: GetConnectionById
): Promise<CreateJobResponse> {
  return createJob(
    new CreateJobRequest({
      accountId,
      jobName: values.define.jobName,
      cronSchedule: values.define.cronSchedule,
      initiateJobRun: values.define.initiateJobRun,
      mappings: [],
      source: toSingleTableAiGenerateJobSource(values),
      destinations: toSingleTableGenerateJobDestinations(
        values,
        getConnectionById
      ),
      workflowOptions: toWorkflowOptions(values.define.workflowSettings),
      syncOptions: toSyncOptions(values),
    }),
    accountId
  );
}

export async function sampleAiGeneratedRecords(
  values: Pick<CreateSingleTableAiGenerateJobFormValues, 'connect' | 'schema'>,
  accountId: string
): Promise<SampleRecord[]> {
  const body = new GetAiGeneratedDataRequest({
    aiConnectionId: values.connect.sourceId,
    count: BigInt(10),
    modelName: values.schema.model,
    userPrompt: values.schema.userPrompt,
    dataConnectionId: values.connect.fkSourceConnectionId,
    table: new DatabaseTable({
      schema: values.schema.schema,
      table: values.schema.table,
    }),
  });

  const res = await fetch(
    `/api/accounts/${accountId}/connections/${values.connect.sourceId}/generate`,
    {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(body),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return (await res.json())?.records ?? [];
}

export async function createNewSingleTableGenerateJob(
  values: CreateSingleTableGenerateJobFormValues,
  accountId: string,
  getConnectionById: GetConnectionById
): Promise<CreateJobResponse> {
  return createJob(
    new CreateJobRequest({
      accountId,
      jobName: values.define.jobName,
      cronSchedule: values.define.cronSchedule,
      initiateJobRun: values.define.initiateJobRun,
      mappings: toSingleGenerateJobMappings(values),
      source: toSingleTableGenerateJobSource(values),
      destinations: toSingleTableGenerateJobDestinations(
        values,
        getConnectionById
      ),
      workflowOptions: toWorkflowOptions(values.define.workflowSettings),
      syncOptions: toSyncOptions(values),
    }),
    accountId
  );
}

function toSingleTableAiGenerateJobSource(
  values: Pick<CreateSingleTableAiGenerateJobFormValues, 'connect' | 'schema'>
): JobSource {
  return new JobSource({
    options: new JobSourceOptions({
      config: {
        case: 'aiGenerate',
        value: new AiGenerateSourceOptions({
          aiConnectionId: values.connect.sourceId,
          modelName: values.schema.model,
          fkSourceConnectionId: values.connect.fkSourceConnectionId,
          userPrompt: values.schema.userPrompt,
          schemas: [
            new AiGenerateSourceSchemaOption({
              schema: values.schema.schema,
              tables: [
                new AiGenerateSourceTableOption({
                  table: values.schema.table,
                  rowCount: BigInt(values.schema.numRows),
                }),
              ],
            }),
          ],
        }),
      },
    }),
  });
}

function toSingleTableGenerateJobSource(
  values: Pick<CreateSingleTableGenerateJobFormValues, 'connect' | 'schema'>
): JobSource {
  const tableSchema =
    values.schema.mappings.length > 0 ? values.schema.mappings[0].schema : null;
  const table =
    values.schema.mappings.length > 0 ? values.schema.mappings[0].table : null;

  return new JobSource({
    options: {
      config: {
        case: 'generate',
        value: new GenerateSourceOptions({
          fkSourceConnectionId: values.connect.fkSourceConnectionId,
          schemas:
            tableSchema && table
              ? [
                  new GenerateSourceSchemaOption({
                    schema: tableSchema,
                    tables: [
                      new GenerateSourceTableOption({
                        rowCount: BigInt(values.schema.numRows),
                        table: table,
                      }),
                    ],
                  }),
                ]
              : [],
        }),
      },
    },
  });
}

export async function createNewSyncJob(
  values: CreateJobFormValues,
  accountId: string,
  getConnectionById: GetConnectionById
): Promise<CreateJobResponse> {
  return createJob(
    new CreateJobRequest({
      accountId,
      jobName: values.define.jobName,
      cronSchedule: values.define.cronSchedule,
      initiateJobRun: values.define.initiateJobRun,
      mappings: toSyncJobMappings(values),
      source: toJobSource(values, getConnectionById),
      destinations: toSyncJobDestinations(values, getConnectionById),
      workflowOptions: toWorkflowOptions(values.define.workflowSettings),
      syncOptions: toSyncOptions(values),
    }),
    accountId
  );
}

function toWorkflowOptions(
  values?: WorkflowSettingsSchema
): WorkflowOptions | undefined {
  if (values?.runTimeout) {
    return new WorkflowOptions({
      runTimeout: convertMinutesToNanoseconds(values.runTimeout),
    });
  }
  return undefined;
}

function toSyncOptions(
  values: Pick<CreateJobFormValues, 'define'>
): ActivityOptions | undefined {
  if (values.define.syncActivityOptions) {
    const formSyncOpts = values.define.syncActivityOptions;
    return new ActivityOptions({
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
  return undefined;
}

function toSingleTableGenerateJobDestinations(
  values: Pick<CreateSingleTableGenerateJobFormValues, 'connect'>,
  getConnectionById: GetConnectionById
): JobDestination[] {
  return [
    new JobDestination({
      connectionId: values.connect.destination.connectionId,
      options: toJobDestinationOptions(
        values.connect.destination,
        getConnectionById(values.connect.destination.connectionId)
      ),
    }),
  ];
}

function toSyncJobDestinations(
  values: Pick<CreateJobFormValues, 'connect'>,
  getConnectionById: GetConnectionById
): JobDestination[] {
  return values.connect.destinations.map((d) => {
    return new JobDestination({
      connectionId: d.connectionId,
      options: toJobDestinationOptions(d, getConnectionById(d.connectionId)),
    });
  });
}

function toJobDestinationOptions(
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
            onConflict: new PostgresOnConflictConfig({
              doNothing: values.destinationOptions.onConflictDoNothing ?? false,
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
            onConflict: new MysqlOnConflictConfig({
              doNothing: values.destinationOptions.onConflictDoNothing ?? false,
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
    case 'mongoConfig': {
      return new JobDestinationOptions({
        config: {
          case: 'mongodbOptions',
          value: new MongoDBDestinationConnectionOptions({}),
        },
      });
    }
    default: {
      return new JobDestinationOptions();
    }
  }
}

function toSingleGenerateJobMappings(
  values: Pick<CreateSingleTableGenerateJobFormValues, 'schema'>
): JobMapping[] {
  return values.schema.mappings.map((m) => {
    return new JobMapping({
      schema: m.schema,
      table: m.table,
      column: m.column,
      transformer: convertJobMappingTransformerFormToJobMappingTransformer(
        m.transformer
      ),
    });
  });
}

function toSyncJobMappings(
  values: Pick<CreateJobFormValues, 'schema'>
): JobMapping[] {
  return values.schema.mappings.map((m) => {
    return new JobMapping({
      schema: m.schema,
      table: m.table,
      column: m.column,
      transformer: convertJobMappingTransformerFormToJobMappingTransformer(
        m.transformer
      ),
    });
  });
}

function toJobSource(
  values: Pick<CreateJobFormValues, 'connect' | 'subset'>,
  getConnectionById: GetConnectionById
): JobSource {
  return new JobSource({
    options: toJobSourceOptions(values, getConnectionById),
  });
}

function toJobSourceOptions(
  values: Pick<CreateJobFormValues, 'connect' | 'subset'>,
  getConnectionById: GetConnectionById
): JobSourceOptions {
  const sourceConnection = getConnectionById(values.connect.sourceId);
  if (!sourceConnection) {
    return new JobSourceOptions();
  }
  switch (sourceConnection.connectionConfig?.config.case) {
    case 'pgConfig':
      return new JobSourceOptions({
        config: {
          case: 'postgres',
          value: new PostgresSourceConnectionOptions({
            connectionId: values.connect.sourceId,
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
            connectionId: values.connect.sourceId,
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
            connectionId: values.connect.sourceId,
          }),
        },
      });
    default:
      throw new Error('unsupported connection type');
  }
}

function toPostgresSourceSchemaOptions(
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

function toMysqlSourceSchemaOptions(
  subsets: SubsetFormValues['subsets']
): MysqlSourceSchemaOption[] {
  const schemaMap = subsets.reduce(
    (map, subset) => {
      if (!map[subset.schema]) {
        map[subset.schema] = new MysqlSourceSchemaOption({
          schema: subset.schema,
          tables: [],
        });
      }
      map[subset.schema].tables.push(
        new MysqlSourceTableOption({
          table: subset.table,
          whereClause: subset.whereClause,
        })
      );
      return map;
    },
    {} as Record<string, MysqlSourceSchemaOption>
  );
  return Object.values(schemaMap);
}

async function createJob(
  input: CreateJobRequest,
  accountId: string
): Promise<CreateJobResponse> {
  const res = await fetch(`/api/accounts/${accountId}/jobs`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CreateJobResponse.fromJson(await res.json());
}

export async function removeJob(
  accountId: string,
  jobId: string
): Promise<void> {
  const res = await fetch(`/api/accounts/${accountId}/jobs/${jobId}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}

export async function isJobNameAvailable(
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

export async function triggerJobRun(
  accountId: string,
  jobId: string
): Promise<void> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${jobId}/create-run`,
    {
      method: 'POST',
      body: JSON.stringify({ jobId }),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}

export async function createJobConnections(
  jobId: string,
  values: DestinationFormValues[],
  connections: Connection[],
  accountId: string
): Promise<CreateJobDestinationConnectionsResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${jobId}/destination-connections`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(
        new CreateJobDestinationConnectionsRequest({
          jobId: jobId,
          destinations: values.map((d) => {
            return new JobDestination({
              connectionId: d.connectionId,
              options: toJobDestinationOptions(
                d,
                connections.find((c) => c.id === d.connectionId)
              ),
            });
          }),
        })
      ),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CreateJobDestinationConnectionsResponse.fromJson(await res.json());
}

export async function updateJobSyncActivityOptions(
  accountId: string,
  jobId: string,
  values: ActivityOptionsSchema
): Promise<SetJobWorkflowOptionsResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${jobId}/syncoptions`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(
        new SetJobSyncOptionsRequest({
          id: jobId,
          syncOptions: new ActivityOptions({
            startToCloseTimeout:
              values.startToCloseTimeout !== undefined &&
              values.startToCloseTimeout > 0
                ? convertMinutesToNanoseconds(values.startToCloseTimeout)
                : undefined,
            scheduleToCloseTimeout:
              values.scheduleToCloseTimeout !== undefined &&
              values.scheduleToCloseTimeout > 0
                ? convertMinutesToNanoseconds(values.scheduleToCloseTimeout)
                : undefined,
            retryPolicy: new RetryPolicy({
              maximumAttempts: values.retryPolicy?.maximumAttempts,
            }),
          }),
        })
      ),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return SetJobWorkflowOptionsResponse.fromJson(await res.json());
}

export async function pauseJob(
  accountId: string,
  jobId: string,
  isPaused: boolean
): Promise<PauseJobResponse> {
  const res = await fetch(`/api/accounts/${accountId}/jobs/${jobId}/pause`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new PauseJobRequest({
        id: jobId,
        pause: isPaused,
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return PauseJobResponse.fromJson(await res.json());
}

export async function updateJobSchedule(
  accountId: string,
  jobId: string,
  cronSchedule?: string
): Promise<UpdateJobScheduleResponse> {
  const res = await fetch(`/api/accounts/${accountId}/jobs/${jobId}/schedule`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new UpdateJobScheduleRequest({
        id: jobId,
        cronSchedule: cronSchedule,
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateJobScheduleResponse.fromJson(await res.json());
}

export async function updateJobWorkflowOptions(
  accountId: string,
  jobId: string,
  values: WorkflowSettingsSchema
): Promise<SetJobWorkflowOptionsResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${jobId}/workflowoptions`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(
        new SetJobWorkflowOptionsRequest({
          id: jobId,
          worfklowOptions: toWorkflowOptions(values),
        })
      ),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return SetJobWorkflowOptionsResponse.fromJson(await res.json());
}

export async function deleteJobConnection(
  accountId: string,
  jobId: string,
  destinationId: string
): Promise<void> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${jobId}/destination-connection/${destinationId}`,
    {
      method: 'DELETE',
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}

export async function setJobConnection(
  accountId: string,
  jobId: string,
  values: DestinationFormValues,
  destinationId: string,
  connection?: Connection
): Promise<UpdateJobDestinationConnectionResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${jobId}/destination-connection`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(
        new UpdateJobDestinationConnectionRequest({
          jobId: jobId,
          connectionId: values.connectionId,
          destinationId: destinationId,
          options: new JobDestinationOptions(
            toJobDestinationOptions(values, connection)
          ),
        })
      ),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateJobDestinationConnectionResponse.fromJson(await res.json());
}

export async function setJobSubsets(
  accountId: string,
  jobId: string,
  values: SubsetFormValues,
  dbType: string
): Promise<SetJobSourceSqlConnectionSubsetsResponse> {
  const schemas =
    dbType == 'mysql'
      ? new JobSourceSqlSubetSchemas({
          schemas: {
            case: 'mysqlSubset',
            value: new MysqlSourceSchemaSubset({
              mysqlSchemas: toMysqlSourceSchemaOptions(values.subsets),
            }),
          },
        })
      : new JobSourceSqlSubetSchemas({
          schemas: {
            case: 'postgresSubset',
            value: new PostgresSourceSchemaSubset({
              postgresSchemas: toPostgresSourceSchemaOptions(values.subsets),
            }),
          },
        });
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${jobId}/source-connection/subsets`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(
        new SetJobSourceSqlConnectionSubsetsRequest({
          id: jobId,
          subsetByForeignKeyConstraints:
            values.subsetOptions.subsetByForeignKeyConstraints,
          schemas,
        })
      ),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return SetJobSourceSqlConnectionSubsetsResponse.fromJson(await res.json());
}
