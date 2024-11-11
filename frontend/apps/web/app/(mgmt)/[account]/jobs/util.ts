import { ValidSubsetConnectionType } from '@/components/jobs/subsets/utils';
import {
  convertMinutesToNanoseconds,
  convertNanosecondsToMinutes,
} from '@/util/util';
import {
  DynamoDBSourceUnmappedTransformConfigFormValues,
  JobMappingFormValues,
  NewDestinationFormValues,
  SchemaFormValues,
  SchemaFormValuesDestinationOptions,
  VirtualForeignConstraintFormValues,
  convertJobMappingTransformerFormToJobMappingTransformer,
  convertJobMappingTransformerToForm,
  toJobSourcePostgresNewColumnAdditionStrategy,
  toNewColumnAdditionStrategy,
} from '@/yup-validations/jobs';
import { Struct, Value } from '@bufbuild/protobuf';
import {
  ActivityOptions,
  AiGenerateSourceOptions,
  AiGenerateSourceSchemaOption,
  AiGenerateSourceTableOption,
  AwsS3DestinationConnectionOptions,
  BatchConfig,
  Connection,
  CreateJobRequest,
  DatabaseTable,
  DynamoDBDestinationConnectionOptions,
  DynamoDBDestinationTableMapping,
  DynamoDBSourceConnectionOptions,
  DynamoDBSourceSchemaSubset,
  DynamoDBSourceTableOption,
  DynamoDBSourceUnmappedTransformConfig,
  GcpCloudStorageDestinationConnectionOptions,
  GenerateBool,
  GenerateSourceOptions,
  GenerateSourceSchemaOption,
  GenerateSourceTableOption,
  GenerateString,
  GetAiGeneratedDataRequest,
  Job,
  JobDestination,
  JobDestinationOptions,
  JobMapping,
  JobMappingTransformer,
  JobSource,
  JobSourceOptions,
  JobSourceSqlSubetSchemas,
  MongoDBDestinationConnectionOptions,
  MongoDBSourceConnectionOptions,
  MssqlDestinationConnectionOptions,
  MssqlOnConflictConfig,
  MssqlSourceConnectionOptions,
  MssqlSourceSchemaOption,
  MssqlSourceSchemaSubset,
  MssqlSourceTableOption,
  MssqlTruncateTableConfig,
  MysqlDestinationConnectionOptions,
  MysqlOnConflictConfig,
  MysqlSourceConnectionOptions,
  MysqlSourceSchemaOption,
  MysqlSourceSchemaSubset,
  MysqlSourceTableOption,
  MysqlTruncateTableConfig,
  Passthrough,
  PostgresDestinationConnectionOptions,
  PostgresOnConflictConfig,
  PostgresSourceConnectionOptions,
  PostgresSourceSchemaOption,
  PostgresSourceSchemaSubset,
  PostgresSourceTableOption,
  PostgresTruncateTableConfig,
  RetryPolicy,
  TransformerConfig,
  ValidateJobMappingsRequest,
  ValidateJobMappingsResponse,
  VirtualForeignConstraint,
  VirtualForeignKey,
  WorkflowOptions,
} from '@neosync/sdk';
import {
  ActivityOptionsSchema,
  ConnectFormValues,
  CreateJobFormValues,
  CreateSingleTableAiGenerateJobFormValues,
  CreateSingleTableGenerateJobFormValues,
  DefineFormValues,
  SingleTableAiConnectFormValues,
  SingleTableAiSchemaFormValues,
  SingleTableConnectFormValues,
  SingleTableEditAiSourceFormValues,
  SingleTableEditSourceFormValues,
  SingleTableSchemaFormValues,
  SubsetFormValues,
  WorkflowSettingsSchema,
} from '../new/job/job-form-validations';

type GetConnectionById = (id: string) => Connection | undefined;

export function getCreateNewSingleTableAiGenerateJobRequest(
  values: CreateSingleTableAiGenerateJobFormValues,
  accountId: string,
  getConnectionById: GetConnectionById
): CreateJobRequest {
  return new CreateJobRequest({
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
  });
}

export function getSampleAiGeneratedRecordsRequest(
  values: Pick<CreateSingleTableAiGenerateJobFormValues, 'connect' | 'schema'>
): GetAiGeneratedDataRequest {
  return new GetAiGeneratedDataRequest({
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
}
export function getSampleEditAiGeneratedRecordsRequest(
  values: SingleTableEditAiSourceFormValues
): GetAiGeneratedDataRequest {
  return new GetAiGeneratedDataRequest({
    aiConnectionId: values.source.sourceId,
    count: BigInt(10),
    modelName: values.schema.model,
    userPrompt: values.schema.userPrompt,
    dataConnectionId: values.source.fkSourceConnectionId,
    table: new DatabaseTable({
      schema: values.schema.schema,
      table: values.schema.table,
    }),
  });
}

export function getCreateNewSingleTableGenerateJobRequest(
  values: CreateSingleTableGenerateJobFormValues,
  accountId: string,
  getConnectionById: GetConnectionById
): CreateJobRequest {
  return new CreateJobRequest({
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
  });
}

export function fromStructToRecord(struct: Struct): Record<string, unknown> {
  return Object.entries(struct.fields).reduce(
    (output, [key, value]) => {
      output[key] = handleValue(value);
      return output;
    },
    {} as Record<string, unknown>
  );
}

function handleValue(value: Value): unknown {
  switch (value.kind.case) {
    case 'structValue': {
      return fromStructToRecord(value.kind.value);
    }
    case 'listValue': {
      return value.kind.value.values.map((val) => handleValue(val));
    }
    default:
      return value.kind.value;
  }
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
          generateBatchSize: BigInt(values.schema.generateBatchSize),
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
export function toSingleTableEditAiGenerateJobSource(
  values: SingleTableEditAiSourceFormValues
): JobSource {
  return new JobSource({
    options: new JobSourceOptions({
      config: {
        case: 'aiGenerate',
        value: new AiGenerateSourceOptions({
          aiConnectionId: values.source.sourceId,
          modelName: values.schema.model,
          fkSourceConnectionId: values.source.fkSourceConnectionId,
          userPrompt: values.schema.userPrompt,
          generateBatchSize: BigInt(values.schema.generateBatchSize),
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

export function toSingleTableEditGenerateJobSource(
  values: SingleTableEditSourceFormValues
): JobSource {
  const schema = values.mappings.length > 0 ? values.mappings[0].schema : null;
  const table = values.mappings.length > 0 ? values.mappings[0].table : null;
  return new JobSource({
    options: new JobSourceOptions({
      config: {
        case: 'generate',
        value: new GenerateSourceOptions({
          fkSourceConnectionId: values.source.fkSourceConnectionId,
          schemas:
            schema && table
              ? [
                  new GenerateSourceSchemaOption({
                    schema: schema,
                    tables: [
                      new GenerateSourceTableOption({
                        table: table,
                        rowCount: BigInt(values.numRows),
                      }),
                    ],
                  }),
                ]
              : [],
        }),
      },
    }),
  });
}

export function getCreateNewSyncJobRequest(
  values: CreateJobFormValues,
  accountId: string,
  getConnectionById: GetConnectionById
): CreateJobRequest {
  const dstOptRecord = values.schema.destinationOptions.reduce(
    (record, dstOpt) => {
      record[dstOpt.destinationId] = dstOpt;
      return record;
    },
    {} as Record<string, SchemaFormValuesDestinationOptions>
  );
  return new CreateJobRequest({
    accountId,
    jobName: values.define.jobName,
    cronSchedule: values.define.cronSchedule,
    initiateJobRun: values.define.initiateJobRun,
    mappings: toSyncJobMappings(values),
    virtualForeignKeys: toSyncVirtualForeignKeys(values),
    source: toJobSource(values, getConnectionById),
    destinations: toSyncJobDestinations(
      {
        connect: {
          ...values.connect,
          destinations: values.connect.destinations.map((d) => {
            const opt = dstOptRecord[d.connectionId];
            return {
              ...d,
              destinationOptions: {
                ...d.destinationOptions,
                ...opt,
              },
            };
          }),
        },
      },
      getConnectionById
    ),
    workflowOptions: toWorkflowOptions(values.define.workflowSettings),
    syncOptions: toSyncOptions(values),
  });
}

export function toWorkflowOptions(
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

export function toJobDestinationOptions(
  values: NewDestinationFormValues,
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
                values.destinationOptions.postgres?.truncateBeforeInsert ??
                false,
              cascade:
                values.destinationOptions.postgres?.truncateCascade ?? false,
            }),
            onConflict: new PostgresOnConflictConfig({
              doNothing:
                values.destinationOptions.postgres?.onConflictDoNothing ??
                false,
            }),
            initTableSchema:
              values.destinationOptions.postgres?.initTableSchema,
            skipForeignKeyViolations:
              values.destinationOptions.postgres?.skipForeignKeyViolations,
            maxInFlight: values.destinationOptions.postgres?.maxInFlight,
            batch: new BatchConfig({
              ...values.destinationOptions.postgres?.batch,
            }),
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
                values.destinationOptions.mysql?.truncateBeforeInsert ?? false,
            }),
            onConflict: new MysqlOnConflictConfig({
              doNothing:
                values.destinationOptions.mysql?.onConflictDoNothing ?? false,
            }),
            initTableSchema: values.destinationOptions.mysql?.initTableSchema,
            skipForeignKeyViolations:
              values.destinationOptions.mysql?.skipForeignKeyViolations,
            maxInFlight: values.destinationOptions.mysql?.maxInFlight,
            batch: new BatchConfig({
              ...values.destinationOptions.mysql?.batch,
            }),
          }),
        },
      });
    }
    case 'awsS3Config': {
      return new JobDestinationOptions({
        config: {
          case: 'awsS3Options',
          value: new AwsS3DestinationConnectionOptions({
            storageClass: values.destinationOptions.awss3?.storageClass,
            timeout: values.destinationOptions.awss3?.timeout,
            maxInFlight: values.destinationOptions.awss3?.maxInFlight,
            batch: new BatchConfig({
              ...values.destinationOptions.awss3?.batch,
            }),
          }),
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
    case 'gcpCloudstorageConfig': {
      return new JobDestinationOptions({
        config: {
          case: 'gcpCloudstorageOptions',
          value: new GcpCloudStorageDestinationConnectionOptions({}),
        },
      });
    }
    case 'dynamodbConfig': {
      return new JobDestinationOptions({
        config: {
          case: 'dynamodbOptions',
          value: new DynamoDBDestinationConnectionOptions({
            tableMappings:
              values.destinationOptions.dynamodb?.tableMappings.map(
                (tm) =>
                  new DynamoDBDestinationTableMapping({
                    sourceTable: tm.sourceTable,
                    destinationTable: tm.destinationTable,
                  })
              ),
          }),
        },
      });
    }
    case 'mssqlConfig': {
      return new JobDestinationOptions({
        config: {
          case: 'mssqlOptions',
          value: new MssqlDestinationConnectionOptions({
            truncateTable: new MssqlTruncateTableConfig({
              truncateBeforeInsert:
                values.destinationOptions.mssql?.truncateBeforeInsert ?? false,
            }),
            onConflict: new MssqlOnConflictConfig({
              doNothing:
                values.destinationOptions.mssql?.onConflictDoNothing ?? false,
            }),
            initTableSchema: values.destinationOptions.mssql?.initTableSchema,
            skipForeignKeyViolations:
              values.destinationOptions.mssql?.skipForeignKeyViolations,
            maxInFlight: values.destinationOptions.mssql?.maxInFlight,
            batch: new BatchConfig({
              ...values.destinationOptions.mssql?.batch,
            }),
          }),
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

function toSyncVirtualForeignKeys(
  values: Pick<CreateJobFormValues, 'schema'>
): VirtualForeignConstraint[] {
  return (
    values.schema.virtualForeignKeys?.map((v) => {
      return new VirtualForeignConstraint({
        schema: v.schema,
        table: v.table,
        columns: v.columns,
        foreignKey: new VirtualForeignKey({
          schema: v.foreignKey.schema,
          table: v.foreignKey.table,
          columns: v.foreignKey.columns,
        }),
      });
    }) || []
  );
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
            newColumnAdditionStrategy:
              toJobSourcePostgresNewColumnAdditionStrategy(
                values.connect.sourceOptions.postgres?.newColumnAdditionStrategy
              ),
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
              values.connect.sourceOptions.mysql?.haltOnNewColumnAddition ??
              false,
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
    case 'dynamodbConfig': {
      return new JobSourceOptions({
        config: {
          case: 'dynamodb',
          value: new DynamoDBSourceConnectionOptions({
            connectionId: values.connect.sourceId,
            tables: toDynamoDbSourceTableOptions(values.subset?.subsets ?? []),
            unmappedTransforms: toDynamoDbSourceUnmappedOptions(
              values.connect.sourceOptions.dynamodb?.unmappedTransformConfig ??
                getDefaultUnmappedTransformConfig()
            ),
            enableConsistentRead:
              values.connect.sourceOptions.dynamodb?.enableConsistentRead ??
              false,
          }),
        },
      });
    }
    case 'mssqlConfig': {
      return new JobSourceOptions({
        config: {
          case: 'mssql',
          value: new MssqlSourceConnectionOptions({
            connectionId: values.connect.sourceId,
            haltOnNewColumnAddition:
              values.connect.sourceOptions.mssql?.haltOnNewColumnAddition ??
              false,
            subsetByForeignKeyConstraints:
              values.subset?.subsetOptions.subsetByForeignKeyConstraints,
            schemas:
              values.subset?.subsets &&
              toMssqlSourceSchemaOptions(values.subset?.subsets),
          }),
        },
      });
    }
    default:
      throw new Error('unsupported connection type');
  }
}

export function getDefaultUnmappedTransformConfig(): DynamoDBSourceUnmappedTransformConfigFormValues {
  return {
    boolean: convertJobMappingTransformerToForm(
      new JobMappingTransformer({
        config: new TransformerConfig({
          config: {
            case: 'generateBoolConfig',
            value: new GenerateBool(),
          },
        }),
      })
    ),
    byte: convertJobMappingTransformerToForm(
      new JobMappingTransformer({
        config: new TransformerConfig({
          config: {
            case: 'passthroughConfig',
            value: new Passthrough(),
          },
        }),
      })
    ),
    n: convertJobMappingTransformerToForm(
      new JobMappingTransformer({
        config: new TransformerConfig({
          config: {
            case: 'passthroughConfig',
            value: new Passthrough(),
          },
        }),
      })
    ),
    s: convertJobMappingTransformerToForm(
      new JobMappingTransformer({
        config: new TransformerConfig({
          config: {
            case: 'generateStringConfig',
            value: new GenerateString({ min: BigInt(1), max: BigInt(100) }),
          },
        }),
      })
    ),
  };
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

function toMssqlSourceSchemaOptions(
  subsets: SubsetFormValues['subsets']
): MssqlSourceSchemaOption[] {
  const schemaMap = subsets.reduce(
    (map, subset) => {
      if (!map[subset.schema]) {
        map[subset.schema] = new MssqlSourceSchemaOption({
          schema: subset.schema,
          tables: [],
        });
      }
      map[subset.schema].tables.push(
        new MssqlSourceTableOption({
          table: subset.table,
          whereClause: subset.whereClause,
        })
      );
      return map;
    },
    {} as Record<string, MssqlSourceSchemaOption>
  );
  return Object.values(schemaMap);
}

function toDynamoDbSourceTableOptions(
  subsets: SubsetFormValues['subsets']
): DynamoDBSourceTableOption[] {
  return subsets.map(
    (ss) =>
      new DynamoDBSourceTableOption({
        table: ss.table,
        whereClause: ss.whereClause,
      })
  );
}

function toDynamoDbSourceUnmappedOptions(
  unmappedTransformConfig: DynamoDBSourceUnmappedTransformConfigFormValues
): DynamoDBSourceUnmappedTransformConfig {
  return new DynamoDBSourceUnmappedTransformConfig({
    b: convertJobMappingTransformerFormToJobMappingTransformer(
      unmappedTransformConfig.byte
    ),
    boolean: convertJobMappingTransformerFormToJobMappingTransformer(
      unmappedTransformConfig.boolean
    ),
    n: convertJobMappingTransformerFormToJobMappingTransformer(
      unmappedTransformConfig.n
    ),
    s: convertJobMappingTransformerFormToJobMappingTransformer(
      unmappedTransformConfig.s
    ),
  });
}

export function toDynamoDbSourceUnmappedOptionsFormValues(
  ut: DynamoDBSourceUnmappedTransformConfig | undefined
): DynamoDBSourceUnmappedTransformConfigFormValues {
  if (!ut) {
    return getDefaultUnmappedTransformConfig();
  }
  return {
    boolean: convertJobMappingTransformerToForm(
      ut.boolean ||
        new JobMappingTransformer({
          config: new TransformerConfig({
            config: {
              case: 'generateBoolConfig',
              value: new GenerateBool(),
            },
          }),
        })
    ),
    byte: convertJobMappingTransformerToForm(
      ut.b ||
        new JobMappingTransformer({
          config: new TransformerConfig({
            config: {
              case: 'passthroughConfig',
              value: new Passthrough(),
            },
          }),
        })
    ),
    n: convertJobMappingTransformerToForm(
      ut.n ||
        new JobMappingTransformer({
          config: new TransformerConfig({
            config: {
              case: 'passthroughConfig',
              value: new Passthrough(),
            },
          }),
        })
    ),
    s: convertJobMappingTransformerToForm(
      ut.s ||
        new JobMappingTransformer({
          config: new TransformerConfig({
            config: {
              case: 'generateStringConfig',
              value: new GenerateString({ min: BigInt(1), max: BigInt(100) }),
            },
          }),
        })
    ),
  };
}

export function toActivityOptions(
  values: ActivityOptionsSchema
): ActivityOptions {
  return new ActivityOptions({
    startToCloseTimeout:
      values.startToCloseTimeout !== undefined && values.startToCloseTimeout > 0
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
  });
}

export function toJobSourceSqlSubsetSchemas(
  values: SubsetFormValues,
  dbType: ValidSubsetConnectionType | null
): JobSourceSqlSubetSchemas {
  switch (dbType) {
    case 'mysqlConfig': {
      return new JobSourceSqlSubetSchemas({
        schemas: {
          case: 'mysqlSubset',
          value: new MysqlSourceSchemaSubset({
            mysqlSchemas: toMysqlSourceSchemaOptions(values.subsets),
          }),
        },
      });
    }
    case 'pgConfig': {
      return new JobSourceSqlSubetSchemas({
        schemas: {
          case: 'postgresSubset',
          value: new PostgresSourceSchemaSubset({
            postgresSchemas: toPostgresSourceSchemaOptions(values.subsets),
          }),
        },
      });
    }
    case 'dynamodbConfig': {
      return new JobSourceSqlSubetSchemas({
        schemas: {
          case: 'dynamodbSubset',
          value: new DynamoDBSourceSchemaSubset({
            tables: toDynamoDbSourceTableOptions(values.subsets),
          }),
        },
      });
    }
    case 'mssqlConfig': {
      return new JobSourceSqlSubetSchemas({
        schemas: {
          case: 'mssqlSubset',
          value: new MssqlSourceSchemaSubset({
            mssqlSchemas: toMssqlSourceSchemaOptions(values.subsets),
          }),
        },
      });
    }
    default: {
      return new JobSourceSqlSubetSchemas();
    }
  }
}

export function setDefaultNewJobFormValues(
  storage: Storage,
  job: Job,
  sessionId: string
): void {
  setDefaultDefineFormValues(storage, job, sessionId);
  setDefaultConnectFormValues(storage, job, sessionId);
  setDefaultSchemaFormValues(storage, job, sessionId);
  setDefaultSubsetFormValues(storage, job, sessionId);
}

function setDefaultDefineFormValues(
  storage: Storage,
  job: Job,
  sessionPrefix: string
): void {
  const values: DefineFormValues = {
    jobName: `${job.name}-copy`,
    cronSchedule: job.cronSchedule,
    initiateJobRun: false,
    syncActivityOptions: job.syncOptions
      ? {
          retryPolicy: job.syncOptions.retryPolicy,
          scheduleToCloseTimeout: job.syncOptions.scheduleToCloseTimeout
            ? convertNanosecondsToMinutes(
                job.syncOptions.scheduleToCloseTimeout
              )
            : undefined,
          startToCloseTimeout: job.syncOptions.startToCloseTimeout
            ? convertNanosecondsToMinutes(job.syncOptions.startToCloseTimeout)
            : undefined,
        }
      : undefined,
    workflowSettings: job.workflowOptions
      ? {
          runTimeout: job.workflowOptions.runTimeout
            ? convertNanosecondsToMinutes(job.workflowOptions.runTimeout)
            : undefined,
        }
      : undefined,
  };
  storage.setItem(
    getNewJobSessionKeys(sessionPrefix).global.define,
    JSON.stringify(values)
  );
}

function setDefaultConnectFormValues(
  storage: Storage,
  job: Job,
  sessionPrefix: string
): void {
  const sessionKeys = getNewJobSessionKeys(sessionPrefix);
  switch (job.source?.options?.config.case) {
    case 'aiGenerate': {
      const values: SingleTableAiConnectFormValues = {
        sourceId: job.source.options.config.value.aiConnectionId,
        fkSourceConnectionId:
          job.source.options.config.value.fkSourceConnectionId ?? '',
        destination:
          job.destinations.length > 0
            ? getDefaultDestinationFormValues(job.destinations[0])
            : {
                connectionId: '',
                destinationOptions: {},
              },
      };
      storage.setItem(sessionKeys.aigenerate.connect, JSON.stringify(values));
      return;
    }
    case 'generate': {
      const values: SingleTableConnectFormValues = {
        fkSourceConnectionId:
          job.source.options.config.value.fkSourceConnectionId ?? '',
        destination:
          job.destinations.length > 0
            ? getDefaultDestinationFormValues(job.destinations[0])
            : {
                connectionId: '',
                destinationOptions: {},
              },
      };

      storage.setItem(sessionKeys.generate.connect, JSON.stringify(values));
      return;
    }
    case 'mongodb': {
      const values: ConnectFormValues = {
        sourceId: job.source.options.config.value.connectionId,
        sourceOptions: {},
        destinations: job.destinations.map((dest) =>
          getDefaultDestinationFormValues(dest)
        ),
      };

      storage.setItem(sessionKeys.dataSync.connect, JSON.stringify(values));
      return;
    }
    case 'dynamodb': {
      const defaultUnmappedConfig = getDefaultUnmappedTransformConfig();
      const values: ConnectFormValues = {
        sourceId: job.source.options.config.value.connectionId,
        sourceOptions: {
          dynamodb: {
            unmappedTransformConfig: {
              byte: job.source.options.config.value.unmappedTransforms?.b
                ? convertJobMappingTransformerToForm(
                    job.source.options.config.value.unmappedTransforms.b
                  )
                : defaultUnmappedConfig.byte,
              boolean: job.source.options.config.value.unmappedTransforms
                ?.boolean
                ? convertJobMappingTransformerToForm(
                    job.source.options.config.value.unmappedTransforms.boolean
                  )
                : defaultUnmappedConfig.boolean,
              n: job.source.options.config.value.unmappedTransforms?.n
                ? convertJobMappingTransformerToForm(
                    job.source.options.config.value.unmappedTransforms.n
                  )
                : defaultUnmappedConfig.n,
              s: job.source.options.config.value.unmappedTransforms?.s
                ? convertJobMappingTransformerToForm(
                    job.source.options.config.value.unmappedTransforms.s
                  )
                : defaultUnmappedConfig.s,
            },
            enableConsistentRead:
              job.source.options.config.value.enableConsistentRead,
          },
        },
        destinations: job.destinations.map((dest) =>
          getDefaultDestinationFormValues(dest)
        ),
      };
      storage.setItem(sessionKeys.dataSync.connect, JSON.stringify(values));
      return;
    }
    case 'mysql': {
      const values: ConnectFormValues = {
        sourceId: job.source.options.config.value.connectionId,
        sourceOptions: {
          mysql: {
            haltOnNewColumnAddition:
              job.source.options.config.value.haltOnNewColumnAddition,
          },
        },
        destinations: job.destinations.map((dest) =>
          getDefaultDestinationFormValues(dest)
        ),
      };

      storage.setItem(sessionKeys.dataSync.connect, JSON.stringify(values));
      return;
    }
    case 'postgres': {
      const values: ConnectFormValues = {
        sourceId: job.source.options.config.value.connectionId,
        sourceOptions: {
          postgres: {
            newColumnAdditionStrategy: toNewColumnAdditionStrategy(
              job.source.options.config.value.newColumnAdditionStrategy
            ),
          },
        },
        destinations: job.destinations.map((dest) =>
          getDefaultDestinationFormValues(dest)
        ),
      };

      storage.setItem(sessionKeys.dataSync.connect, JSON.stringify(values));
      return;
    }
    case 'mssql': {
      const values: ConnectFormValues = {
        sourceId: job.source.options.config.value.connectionId,
        sourceOptions: {
          mssql: {
            haltOnNewColumnAddition:
              job.source.options.config.value.haltOnNewColumnAddition,
          },
        },
        destinations: job.destinations.map((dest) =>
          getDefaultDestinationFormValues(dest)
        ),
      };

      storage.setItem(sessionKeys.dataSync.connect, JSON.stringify(values));
      return;
    }
  }
}

function setDefaultSchemaFormValues(
  storage: Storage,
  job: Job,
  sessionPrefix: string
): void {
  const sessionKeys = getNewJobSessionKeys(sessionPrefix);
  switch (job.source?.options?.config.case) {
    case 'aiGenerate': {
      const numRows = getSingleTableAiGenerateNumRows(
        job.source.options.config.value
      );
      let genBatchSize = 10;
      if (job.source.options.config.value.generateBatchSize) {
        genBatchSize = Number(
          job.source.options.config.value.generateBatchSize
        );
      } else {
        // batch size has not been set by the user. Set it to our default of 10, or num rows, whichever is lower
        genBatchSize = Math.min(genBatchSize, numRows);
      }

      const values: SingleTableAiSchemaFormValues = {
        numRows,
        generateBatchSize: genBatchSize,
        userPrompt: job.source.options.config.value.userPrompt,
        model: job.source.options.config.value.modelName,
        ...getSingleTableAiSchemaTable(job.source.options.config.value),
      };

      storage.setItem(sessionKeys.aigenerate.schema, JSON.stringify(values));
      return;
    }
    case 'generate': {
      const values: SingleTableSchemaFormValues = {
        numRows: getSingleTableGenerateNumRows(job.source.options.config.value),
        mappings: job.mappings.map((mapping) => {
          return {
            ...mapping,
            transformer: mapping.transformer
              ? convertJobMappingTransformerToForm(mapping.transformer)
              : convertJobMappingTransformerToForm(new JobMappingTransformer()),
          };
        }),
      };

      storage.setItem(sessionKeys.generate.schema, JSON.stringify(values));
      return;
    }
    case 'mysql':
    case 'mongodb':
    case 'postgres':
    case 'mssql': {
      const values: SchemaFormValues = {
        destinationOptions: [],
        connectionId: job.source.options.config.value.connectionId,
        mappings: job.mappings.map((mapping) => {
          return {
            ...mapping,
            transformer: mapping.transformer
              ? convertJobMappingTransformerToForm(mapping.transformer)
              : convertJobMappingTransformerToForm(new JobMappingTransformer()),
          };
        }),
        virtualForeignKeys: job.virtualForeignKeys.map((v) => {
          return {
            ...v,
            foreignKey: {
              schema: v.foreignKey?.schema ?? '',
              table: v.foreignKey?.table ?? '',
              columns: v.foreignKey?.columns ?? [],
            },
          };
        }),
      };

      storage.setItem(sessionKeys.dataSync.schema, JSON.stringify(values));
      return;
    }
    case 'dynamodb': {
      const values: SchemaFormValues = {
        destinationOptions: job.destinations.map((dest) => {
          if (dest.options?.config.case !== 'dynamodbOptions') {
            return { destinationId: dest.id };
          }
          return {
            destinationId: dest.id,
            dynamoDb: {
              tableMappings: dest.options.config.value.tableMappings.map(
                (tm) => ({
                  sourceTable: tm.sourceTable,
                  destinationTable: tm.destinationTable,
                })
              ),
            },
          };
        }),
        connectionId: job.source.options.config.value.connectionId,
        mappings: job.mappings.map((mapping) => {
          return {
            ...mapping,
            transformer: mapping.transformer
              ? convertJobMappingTransformerToForm(mapping.transformer)
              : convertJobMappingTransformerToForm(new JobMappingTransformer()),
          };
        }),
        virtualForeignKeys: job.virtualForeignKeys.map((v) => {
          return {
            ...v,
            foreignKey: {
              schema: v.foreignKey?.schema ?? '',
              table: v.foreignKey?.table ?? '',
              columns: v.foreignKey?.columns ?? [],
            },
          };
        }),
      };

      storage.setItem(sessionKeys.dataSync.schema, JSON.stringify(values));
      return;
    }
  }
}

function setDefaultSubsetFormValues(
  storage: Storage,
  job: Job,
  sessionPrefix: string
): void {
  switch (job.source?.options?.config.case) {
    case 'postgres':
    case 'mysql':
    case 'mssql': {
      const values: SubsetFormValues = {
        subsets: job.source.options.config.value.schemas.flatMap(
          (schema): SubsetFormValues['subsets'] => {
            return schema.tables.map((table) => {
              return {
                schema: schema.schema,
                table: table.table,
                whereClause: table.whereClause,
              };
            });
          }
        ),
        subsetOptions: {
          subsetByForeignKeyConstraints:
            job.source.options.config.value.subsetByForeignKeyConstraints,
        },
      };
      storage.setItem(
        getNewJobSessionKeys(sessionPrefix).dataSync.subset,
        JSON.stringify(values)
      );
      return;
    }
  }
}

export function getSingleTableAiGenerateNumRows(
  sourceOpts: AiGenerateSourceOptions
): number {
  const srcSchemas = sourceOpts.schemas;
  if (srcSchemas.length > 0) {
    const tables = srcSchemas[0].tables;
    if (tables.length > 0) {
      return Number(tables[0].rowCount); // this will be an issue if the number is bigger than what js allows
    }
  }
  return 0;
}

export function getSingleTableAiSchemaTable(
  sourceOpts: AiGenerateSourceOptions
): { schema: string; table: string } {
  const srcSchemas = sourceOpts.schemas;
  if (srcSchemas.length > 0) {
    const tables = srcSchemas[0].tables;
    if (tables.length > 0) {
      return {
        schema: srcSchemas[0].schema,
        table: tables[0].table,
      };
    }
    return { schema: srcSchemas[0].schema, table: '' };
  }
  return { schema: '', table: '' };
}

export function getSingleTableGenerateNumRows(
  sourceOpts: GenerateSourceOptions
): number {
  const srcSchemas = sourceOpts.schemas;
  if (srcSchemas.length > 0) {
    const tables = srcSchemas[0].tables;
    if (tables.length > 0) {
      return Number(tables[0].rowCount); // this will be an issue if the number is bigger than what js allows
    }
  }
  return 0;
}

export function getDefaultDestinationFormValues(
  d: JobDestination
): NewDestinationFormValues {
  switch (d.options?.config.case) {
    case 'postgresOptions': {
      return {
        connectionId: d.connectionId,
        destinationOptions: {
          postgres: {
            truncateBeforeInsert:
              d.options.config.value.truncateTable?.truncateBeforeInsert ??
              false,
            truncateCascade:
              d.options.config.value.truncateTable?.cascade ?? false,
            initTableSchema: d.options.config.value.initTableSchema ?? false,
            onConflictDoNothing:
              d.options.config.value.onConflict?.doNothing ?? false,
            skipForeignKeyViolations:
              d.options.config.value.skipForeignKeyViolations ?? false,
            maxInFlight: d.options.config.value.maxInFlight,
            batch: {
              count: d.options.config.value.batch?.count,
              period: d.options.config.value.batch?.period,
            },
          },
        },
      };
    }
    case 'mysqlOptions': {
      return {
        connectionId: d.connectionId,
        destinationOptions: {
          mysql: {
            truncateBeforeInsert:
              d.options.config.value.truncateTable?.truncateBeforeInsert ??
              false,
            initTableSchema: d.options.config.value.initTableSchema ?? false,
            onConflictDoNothing:
              d.options.config.value.onConflict?.doNothing ?? false,
            skipForeignKeyViolations:
              d.options.config.value.skipForeignKeyViolations ?? false,
            maxInFlight: d.options.config.value.maxInFlight,
            batch: {
              count: d.options.config.value.batch?.count,
              period: d.options.config.value.batch?.period,
            },
          },
        },
      };
    }
    case 'dynamodbOptions': {
      return {
        connectionId: d.connectionId,
        destinationOptions: {
          dynamodb: {
            tableMappings: d.options.config.value.tableMappings.map((tm) => ({
              sourceTable: tm.sourceTable,
              destinationTable: tm.destinationTable,
            })),
          },
        },
      };
    }
    case 'mssqlOptions': {
      return {
        connectionId: d.connectionId,
        destinationOptions: {
          mssql: {
            truncateBeforeInsert:
              d.options.config.value.truncateTable?.truncateBeforeInsert ??
              false,
            initTableSchema: d.options.config.value.initTableSchema ?? false,
            onConflictDoNothing:
              d.options.config.value.onConflict?.doNothing ?? false,
            skipForeignKeyViolations:
              d.options.config.value.skipForeignKeyViolations ?? false,
            maxInFlight: d.options.config.value.maxInFlight,
            batch: {
              count: d.options.config.value.batch?.count,
              period: d.options.config.value.batch?.period,
            },
          },
        },
      };
    }
    case 'awsS3Options': {
      return {
        connectionId: d.connectionId,
        destinationOptions: {
          awss3: {
            storageClass: d.options.config.value.storageClass,
            maxInFlight: d.options.config.value.maxInFlight,
            timeout: d.options.config.value.timeout,
            batch: {
              count: d.options.config.value.batch?.count,
              period: d.options.config.value.batch?.period,
            },
          },
        },
      };
    }
    default:
      return {
        connectionId: d.connectionId,
        destinationOptions: {},
      };
  }
}

interface NewJobSessionKeys {
  global: { define: string };
  dataSync: { connect: string; schema: string; subset: string };
  generate: { connect: string; schema: string };
  aigenerate: { connect: string; schema: string };
}

export function getNewJobSessionKeys(sessionId: string): NewJobSessionKeys {
  return {
    global: {
      define: `${sessionId}-new-job-define`,
    },
    dataSync: {
      connect: `${sessionId}-new-job-connect`,
      schema: `${sessionId}-new-job-schema`,
      subset: `${sessionId}-new-job-subset`,
    },
    generate: {
      connect: `${sessionId}-new-job-single-table-connect`,
      schema: `${sessionId}-new-job-single-table-schema`,
    },
    aigenerate: {
      connect: `${sessionId}-new-job-single-table-ai-connect`,
      schema: `${sessionId}-new-job-single-table-ai-schema`,
    },
  };
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function collectStringLeafs(obj: any): string[] {
  if (typeof obj === 'string') {
    return [obj];
  } else if (typeof obj === 'object' && obj != null) {
    return Object.keys(obj).flatMap((key) => collectStringLeafs(obj[key]));
  }
  return [];
}

export function clearNewJobSession(storage: Storage, sessionId: string): void {
  const keys = collectStringLeafs(getNewJobSessionKeys(sessionId));
  keys.forEach((key) => storage.removeItem(key));
}

export async function validateJobMapping(
  connectionId: string,
  formMappings: JobMappingFormValues[],
  accountId: string,
  virtualForeignKeys: VirtualForeignConstraintFormValues[],
  validate: (
    req: ValidateJobMappingsRequest
  ) => Promise<ValidateJobMappingsResponse>
): Promise<ValidateJobMappingsResponse> {
  const body = new ValidateJobMappingsRequest({
    accountId,
    mappings: formMappings.map((m) => {
      return new JobMapping({
        schema: m.schema,
        table: m.table,
        column: m.column,
        transformer: m.transformer.config.case
          ? convertJobMappingTransformerFormToJobMappingTransformer(
              m.transformer
            )
          : new JobMappingTransformer({
              config: new TransformerConfig({
                config: {
                  case: 'passthroughConfig',
                  value: {},
                },
              }),
            }),
      });
    }),
    virtualForeignKeys: virtualForeignKeys.map((v) => {
      return new VirtualForeignConstraint({
        schema: v.schema,
        table: v.table,
        columns: v.columns,
        foreignKey: new VirtualForeignKey({
          schema: v.foreignKey.schema,
          table: v.foreignKey.table,
          columns: v.foreignKey.columns,
        }),
      });
    }),
    connectionId: connectionId,
  });

  return validate(body);
}
