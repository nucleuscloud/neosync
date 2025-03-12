import { ValidSubsetConnectionType } from '@/components/jobs/subsets/utils';
import {
  convertMinutesToNanoseconds,
  convertNanosecondsToMinutes,
} from '@/util/util';
import {
  AwsS3DestinationOptionsFormValues,
  convertJobMappingTransformerFormToJobMappingTransformer,
  convertJobMappingTransformerToForm,
  DestinationOptionsFormValues,
  DynamoDBSourceUnmappedTransformConfigFormValues,
  JobMappingFormValues,
  MssqlDbDestinationOptionsFormValues,
  MysqlDbDestinationOptionsFormValues,
  NewDestinationFormValues,
  PostgresDbDestinationOptionsFormValues,
  SchemaFormValues,
  SchemaFormValuesDestinationOptions,
  toColumnRemovalStrategy,
  toJobSourceMssqlColumnRemovalStrategy,
  toJobSourceMysqlColumnRemovalStrategy,
  toJobSourcePostgresColumnRemovalStrategy,
  toJobSourcePostgresNewColumnAdditionStrategy,
  toNewColumnAdditionStrategy,
  VirtualForeignConstraintFormValues,
} from '@/yup-validations/jobs';
import { create } from '@bufbuild/protobuf';
import {
  ActivityOptions,
  ActivityOptionsSchema,
  AiGenerateSourceOptions,
  AiGenerateSourceOptionsSchema,
  AiGenerateSourceSchemaOptionSchema,
  AiGenerateSourceTableOptionSchema,
  AwsS3DestinationConnectionOptions_StorageClass,
  AwsS3DestinationConnectionOptionsSchema,
  BatchConfigSchema,
  Connection,
  CreateJobDestination,
  CreateJobDestinationSchema,
  CreateJobRequest,
  CreateJobRequestSchema,
  DatabaseTableSchema,
  DynamoDBDestinationConnectionOptionsSchema,
  DynamoDBDestinationTableMappingSchema,
  DynamoDBSourceConnectionOptionsSchema,
  DynamoDBSourceSchemaSubsetSchema,
  DynamoDBSourceTableOption,
  DynamoDBSourceTableOptionSchema,
  DynamoDBSourceUnmappedTransformConfig,
  DynamoDBSourceUnmappedTransformConfigSchema,
  GcpCloudStorageDestinationConnectionOptionsSchema,
  GenerateBoolSchema,
  GenerateSourceOptions,
  GenerateSourceOptionsSchema,
  GenerateSourceSchemaOptionSchema,
  GenerateSourceTableOptionSchema,
  GenerateStringSchema,
  GetAiGeneratedDataRequest,
  GetAiGeneratedDataRequestSchema,
  Job,
  JobDestination,
  JobDestinationOptions,
  JobDestinationOptionsSchema,
  JobMapping,
  JobMappingSchema,
  JobMappingTransformerSchema,
  JobSource,
  JobSourceOptions,
  JobSourceOptionsSchema,
  JobSourceSchema,
  JobSourceSqlSubetSchemas,
  JobSourceSqlSubetSchemasSchema,
  MongoDBDestinationConnectionOptionsSchema,
  MongoDBSourceConnectionOptionsSchema,
  MssqlDestinationConnectionOptionsSchema,
  MssqlOnConflictConfigSchema,
  MssqlSourceConnectionOptionsSchema,
  MssqlSourceSchemaOption,
  MssqlSourceSchemaOptionSchema,
  MssqlSourceSchemaSubsetSchema,
  MssqlSourceTableOptionSchema,
  MssqlTruncateTableConfigSchema,
  MysqlDestinationConnectionOptionsSchema,
  MysqlOnConflictConfig_MysqlOnConflictDoNothingSchema,
  MysqlOnConflictConfig_MysqlOnConflictUpdateSchema,
  MysqlOnConflictConfigSchema,
  MysqlSourceConnectionOptionsSchema,
  MysqlSourceSchemaOption,
  MysqlSourceSchemaOptionSchema,
  MysqlSourceSchemaSubsetSchema,
  MysqlSourceTableOptionSchema,
  MysqlTruncateTableConfigSchema,
  PassthroughSchema,
  PostgresDestinationConnectionOptionsSchema,
  PostgresOnConflictConfig_PostgresOnConflictDoNothingSchema,
  PostgresOnConflictConfig_PostgresOnConflictUpdateSchema,
  PostgresOnConflictConfigSchema,
  PostgresSourceConnectionOptionsSchema,
  PostgresSourceSchemaOption,
  PostgresSourceSchemaOptionSchema,
  PostgresSourceSchemaSubsetSchema,
  PostgresSourceTableOptionSchema,
  PostgresTruncateTableConfigSchema,
  RetryPolicySchema,
  TransformerConfigSchema,
  ValidateJobMappingsRequest,
  ValidateJobMappingsRequestSchema,
  ValidateJobMappingsResponse,
  VirtualForeignConstraint,
  VirtualForeignConstraintSchema,
  VirtualForeignKeySchema,
  WorkflowOptions,
  WorkflowOptionsSchema,
} from '@neosync/sdk';
import { ConnectionConfigCase } from '../connections/util';
import {
  ActivityOptionsFormValues,
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
import { getConnectionIdFromSource } from './[id]/source/components/util';

type GetConnectionById = (id: string) => Connection | undefined;

export function getCreateNewSingleTableAiGenerateJobRequest(
  values: CreateSingleTableAiGenerateJobFormValues,
  accountId: string,
  getConnectionById: GetConnectionById
): CreateJobRequest {
  return create(CreateJobRequestSchema, {
    accountId,
    jobName: values.define.jobName,
    cronSchedule: values.define.cronSchedule,
    initiateJobRun: values.define.initiateJobRun,
    mappings: [],
    source: toSingleTableAiGenerateJobSource(values),
    destinations: toSingleTableGenerateCreateJobDestinations(
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
  return create(GetAiGeneratedDataRequestSchema, {
    aiConnectionId: values.connect.sourceId,
    count: BigInt(Math.min(10, values.schema.numRows)),
    modelName: values.schema.model,
    userPrompt: values.schema.userPrompt,
    dataConnectionId: values.connect.fkSourceConnectionId,
    table: create(DatabaseTableSchema, {
      schema: values.schema.schema,
      table: values.schema.table,
    }),
  });
}
export function getSampleEditAiGeneratedRecordsRequest(
  values: SingleTableEditAiSourceFormValues
): GetAiGeneratedDataRequest {
  return create(GetAiGeneratedDataRequestSchema, {
    aiConnectionId: values.source.sourceId,
    count: BigInt(10),
    modelName: values.schema.model,
    userPrompt: values.schema.userPrompt,
    dataConnectionId: values.source.fkSourceConnectionId,
    table: create(DatabaseTableSchema, {
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
  return create(CreateJobRequestSchema, {
    accountId,
    jobName: values.define.jobName,
    cronSchedule: values.define.cronSchedule,
    initiateJobRun: values.define.initiateJobRun,
    mappings: toSingleGenerateJobMappings(values),
    source: toSingleTableGenerateJobSource(values),
    destinations: toSingleTableGenerateCreateJobDestinations(
      values,
      getConnectionById
    ),
    workflowOptions: toWorkflowOptions(values.define.workflowSettings),
    syncOptions: toSyncOptions(values),
  });
}

function toSingleTableAiGenerateJobSource(
  values: Pick<CreateSingleTableAiGenerateJobFormValues, 'connect' | 'schema'>
): JobSource {
  return create(JobSourceSchema, {
    options: create(JobSourceOptionsSchema, {
      config: {
        case: 'aiGenerate',
        value: create(AiGenerateSourceOptionsSchema, {
          aiConnectionId: values.connect.sourceId,
          modelName: values.schema.model,
          fkSourceConnectionId: values.connect.fkSourceConnectionId,
          userPrompt: values.schema.userPrompt,
          generateBatchSize: BigInt(values.schema.generateBatchSize),
          schemas: [
            create(AiGenerateSourceSchemaOptionSchema, {
              schema: values.schema.schema,
              tables: [
                create(AiGenerateSourceTableOptionSchema, {
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
  return create(JobSourceSchema, {
    options: create(JobSourceOptionsSchema, {
      config: {
        case: 'aiGenerate',
        value: create(AiGenerateSourceOptionsSchema, {
          aiConnectionId: values.source.sourceId,
          modelName: values.schema.model,
          fkSourceConnectionId: values.source.fkSourceConnectionId,
          userPrompt: values.schema.userPrompt,
          generateBatchSize: BigInt(values.schema.generateBatchSize),
          schemas: [
            create(AiGenerateSourceSchemaOptionSchema, {
              schema: values.schema.schema,
              tables: [
                create(AiGenerateSourceTableOptionSchema, {
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

export function toSingleTableGenerateJobSource(
  values: Pick<CreateSingleTableGenerateJobFormValues, 'connect' | 'schema'>
): JobSource {
  const tableSchema =
    values.schema.mappings.length > 0 ? values.schema.mappings[0].schema : null;
  const table =
    values.schema.mappings.length > 0 ? values.schema.mappings[0].table : null;

  return create(JobSourceSchema, {
    options: create(JobSourceOptionsSchema, {
      config: {
        case: 'generate',
        value: create(GenerateSourceOptionsSchema, {
          fkSourceConnectionId: values.connect.fkSourceConnectionId,
          schemas:
            tableSchema && table
              ? [
                  create(GenerateSourceSchemaOptionSchema, {
                    schema: tableSchema,
                    tables: [
                      create(GenerateSourceTableOptionSchema, {
                        rowCount: BigInt(values.schema.numRows),
                        table: table,
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

export function toSingleTableEditGenerateJobSource(
  values: SingleTableEditSourceFormValues
): JobSource {
  const schema = values.mappings.length > 0 ? values.mappings[0].schema : null;
  const table = values.mappings.length > 0 ? values.mappings[0].table : null;
  return create(JobSourceSchema, {
    options: create(JobSourceOptionsSchema, {
      config: {
        case: 'generate',
        value: create(GenerateSourceOptionsSchema, {
          fkSourceConnectionId: values.source.fkSourceConnectionId,
          schemas:
            schema && table
              ? [
                  create(GenerateSourceSchemaOptionSchema, {
                    schema: schema,
                    tables: [
                      create(GenerateSourceTableOptionSchema, {
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
  return create(CreateJobRequestSchema, {
    accountId,
    jobName: values.define.jobName,
    cronSchedule: values.define.cronSchedule,
    initiateJobRun: values.define.initiateJobRun,
    mappings: toSyncJobMappings(values),
    virtualForeignKeys: toSyncVirtualForeignKeys(values),
    source: toJobSource(values, getConnectionById),
    destinations: toSyncCreateJobDestinations(
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
    jobType: {
      jobType: {
        case: 'sync',
        value: {},
      },
    },
  });
}

export function toWorkflowOptions(
  values?: WorkflowSettingsSchema
): WorkflowOptions | undefined {
  if (values?.runTimeout) {
    return create(WorkflowOptionsSchema, {
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
    return create(ActivityOptionsSchema, {
      scheduleToCloseTimeout:
        formSyncOpts.scheduleToCloseTimeout !== undefined
          ? convertMinutesToNanoseconds(formSyncOpts.scheduleToCloseTimeout)
          : undefined,
      startToCloseTimeout:
        formSyncOpts.startToCloseTimeout !== undefined
          ? convertMinutesToNanoseconds(formSyncOpts.startToCloseTimeout)
          : undefined,
      retryPolicy: create(RetryPolicySchema, {
        maximumAttempts: formSyncOpts.retryPolicy?.maximumAttempts,
      }),
    });
  }
  return undefined;
}

function toSingleTableGenerateCreateJobDestinations(
  values: Pick<CreateSingleTableGenerateJobFormValues, 'connect'>,
  getConnectionById: GetConnectionById
): CreateJobDestination[] {
  return [
    create(CreateJobDestinationSchema, {
      connectionId: values.connect.destination.connectionId,
      options: toJobDestinationOptions(
        values.connect.destination,
        getConnectionById(values.connect.destination.connectionId)
      ),
    }),
  ];
}

function toSyncCreateJobDestinations(
  values: Pick<CreateJobFormValues, 'connect'>,
  getConnectionById: GetConnectionById
): CreateJobDestination[] {
  return values.connect.destinations.map((d) => {
    return create(CreateJobDestinationSchema, {
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
    return create(JobDestinationOptionsSchema, {});
  }
  switch (connection.connectionConfig?.config.case) {
    case 'pgConfig': {
      var pgOnConflict = create(PostgresOnConflictConfigSchema, {});
      if (
        values.destinationOptions.postgres?.conflictStrategy
          ?.onConflictDoNothing
      ) {
        pgOnConflict.strategy = {
          case: 'nothing',
          value: create(
            PostgresOnConflictConfig_PostgresOnConflictDoNothingSchema,
            {}
          ),
        };
      } else if (
        values.destinationOptions.postgres?.conflictStrategy?.onConflictDoUpdate
      ) {
        pgOnConflict.strategy = {
          case: 'update',
          value: create(
            PostgresOnConflictConfig_PostgresOnConflictUpdateSchema,
            {}
          ),
        };
      }
      return create(JobDestinationOptionsSchema, {
        config: {
          case: 'postgresOptions',
          value: create(PostgresDestinationConnectionOptionsSchema, {
            truncateTable: create(PostgresTruncateTableConfigSchema, {
              truncateBeforeInsert:
                values.destinationOptions.postgres?.truncateBeforeInsert ??
                false,
              cascade:
                values.destinationOptions.postgres?.truncateCascade ?? false,
            }),
            onConflict: pgOnConflict,
            initTableSchema:
              values.destinationOptions.postgres?.initTableSchema,
            skipForeignKeyViolations:
              values.destinationOptions.postgres?.skipForeignKeyViolations,
            maxInFlight: values.destinationOptions.postgres?.maxInFlight,
            batch: create(BatchConfigSchema, {
              ...values.destinationOptions.postgres?.batch,
            }),
          }),
        },
      });
    }
    case 'mysqlConfig': {
      var onConflict = create(MysqlOnConflictConfigSchema, {});
      if (
        values.destinationOptions.mysql?.conflictStrategy?.onConflictDoNothing
      ) {
        onConflict.strategy = {
          case: 'nothing',
          value: create(
            MysqlOnConflictConfig_MysqlOnConflictDoNothingSchema,
            {}
          ),
        };
      } else if (
        values.destinationOptions.mysql?.conflictStrategy?.onConflictDoUpdate
      ) {
        onConflict.strategy = {
          case: 'update',
          value: create(MysqlOnConflictConfig_MysqlOnConflictUpdateSchema, {}),
        };
      }
      return create(JobDestinationOptionsSchema, {
        config: {
          case: 'mysqlOptions',
          value: create(MysqlDestinationConnectionOptionsSchema, {
            truncateTable: create(MysqlTruncateTableConfigSchema, {
              truncateBeforeInsert:
                values.destinationOptions.mysql?.truncateBeforeInsert ?? false,
            }),
            onConflict: onConflict,
            initTableSchema: values.destinationOptions.mysql?.initTableSchema,
            skipForeignKeyViolations:
              values.destinationOptions.mysql?.skipForeignKeyViolations,
            maxInFlight: values.destinationOptions.mysql?.maxInFlight,
            batch: create(BatchConfigSchema, {
              ...values.destinationOptions.mysql?.batch,
            }),
          }),
        },
      });
    }
    case 'awsS3Config': {
      return create(JobDestinationOptionsSchema, {
        config: {
          case: 'awsS3Options',
          value: create(AwsS3DestinationConnectionOptionsSchema, {
            storageClass: values.destinationOptions.awss3?.storageClass,
            timeout: values.destinationOptions.awss3?.timeout,
            maxInFlight: values.destinationOptions.awss3?.maxInFlight,
            batch: create(BatchConfigSchema, {
              ...values.destinationOptions.awss3?.batch,
            }),
          }),
        },
      });
    }
    case 'mongoConfig': {
      return create(JobDestinationOptionsSchema, {
        config: {
          case: 'mongodbOptions',
          value: create(MongoDBDestinationConnectionOptionsSchema, {}),
        },
      });
    }
    case 'gcpCloudstorageConfig': {
      return create(JobDestinationOptionsSchema, {
        config: {
          case: 'gcpCloudstorageOptions',
          value: create(GcpCloudStorageDestinationConnectionOptionsSchema, {}),
        },
      });
    }
    case 'dynamodbConfig': {
      return create(JobDestinationOptionsSchema, {
        config: {
          case: 'dynamodbOptions',
          value: create(DynamoDBDestinationConnectionOptionsSchema, {
            tableMappings:
              values.destinationOptions.dynamodb?.tableMappings.map((tm) =>
                create(DynamoDBDestinationTableMappingSchema, {
                  sourceTable: tm.sourceTable,
                  destinationTable: tm.destinationTable,
                })
              ),
          }),
        },
      });
    }
    case 'mssqlConfig': {
      return create(JobDestinationOptionsSchema, {
        config: {
          case: 'mssqlOptions',
          value: create(MssqlDestinationConnectionOptionsSchema, {
            truncateTable: create(MssqlTruncateTableConfigSchema, {
              truncateBeforeInsert:
                values.destinationOptions.mssql?.truncateBeforeInsert ?? false,
            }),
            onConflict: create(MssqlOnConflictConfigSchema, {
              doNothing:
                values.destinationOptions.mssql?.onConflictDoNothing ?? false,
            }),
            initTableSchema: values.destinationOptions.mssql?.initTableSchema,
            skipForeignKeyViolations:
              values.destinationOptions.mssql?.skipForeignKeyViolations,
            maxInFlight: values.destinationOptions.mssql?.maxInFlight,
            batch: create(BatchConfigSchema, {
              ...values.destinationOptions.mssql?.batch,
            }),
          }),
        },
      });
    }
    default: {
      return create(JobDestinationOptionsSchema, {});
    }
  }
}

function toSingleGenerateJobMappings(
  values: Pick<CreateSingleTableGenerateJobFormValues, 'schema'>
): JobMapping[] {
  return values.schema.mappings.map((m) => {
    return create(JobMappingSchema, {
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
    return create(JobMappingSchema, {
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
      return create(VirtualForeignConstraintSchema, {
        schema: v.schema,
        table: v.table,
        columns: v.columns,
        foreignKey: create(VirtualForeignKeySchema, {
          schema: v.foreignKey.schema,
          table: v.foreignKey.table,
          columns: v.foreignKey.columns,
        }),
      });
    }) || []
  );
}

export function toJobSource(
  values: Pick<CreateJobFormValues, 'connect' | 'subset'>,
  getConnectionById: GetConnectionById
): JobSource {
  return create(JobSourceSchema, {
    options: toJobSourceOptions(values, getConnectionById),
  });
}

function toJobSourceOptions(
  values: Pick<CreateJobFormValues, 'connect' | 'subset'>,
  getConnectionById: GetConnectionById
): JobSourceOptions {
  const sourceConnection = getConnectionById(values.connect.sourceId);
  if (!sourceConnection) {
    return create(JobSourceOptionsSchema, {});
  }
  switch (sourceConnection.connectionConfig?.config.case) {
    case 'pgConfig':
      return create(JobSourceOptionsSchema, {
        config: {
          case: 'postgres',
          value: create(PostgresSourceConnectionOptionsSchema, {
            connectionId: values.connect.sourceId,
            newColumnAdditionStrategy:
              toJobSourcePostgresNewColumnAdditionStrategy(
                values.connect.sourceOptions.postgres?.newColumnAdditionStrategy
              ),
            columnRemovalStrategy: toJobSourcePostgresColumnRemovalStrategy(
              values.connect.sourceOptions.postgres?.columnRemovalStrategy
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
      return create(JobSourceOptionsSchema, {
        config: {
          case: 'mysql',
          value: create(MysqlSourceConnectionOptionsSchema, {
            connectionId: values.connect.sourceId,
            haltOnNewColumnAddition:
              values.connect.sourceOptions.mysql?.haltOnNewColumnAddition ??
              false,
            columnRemovalStrategy: toJobSourceMysqlColumnRemovalStrategy(
              values.connect.sourceOptions.mysql?.columnRemovalStrategy
            ),
            subsetByForeignKeyConstraints:
              values.subset?.subsetOptions.subsetByForeignKeyConstraints,
            schemas:
              values.subset?.subsets &&
              toMysqlSourceSchemaOptions(values.subset?.subsets),
          }),
        },
      });
    case 'mongoConfig':
      return create(JobSourceOptionsSchema, {
        config: {
          case: 'mongodb',
          value: create(MongoDBSourceConnectionOptionsSchema, {
            connectionId: values.connect.sourceId,
          }),
        },
      });
    case 'dynamodbConfig': {
      return create(JobSourceOptionsSchema, {
        config: {
          case: 'dynamodb',
          value: create(DynamoDBSourceConnectionOptionsSchema, {
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
      return create(JobSourceOptionsSchema, {
        config: {
          case: 'mssql',
          value: create(MssqlSourceConnectionOptionsSchema, {
            connectionId: values.connect.sourceId,
            haltOnNewColumnAddition:
              values.connect.sourceOptions.mssql?.haltOnNewColumnAddition ??
              false,
            columnRemovalStrategy: toJobSourceMssqlColumnRemovalStrategy(
              values.connect.sourceOptions.mssql?.columnRemovalStrategy
            ),
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
      create(JobMappingTransformerSchema, {
        config: create(TransformerConfigSchema, {
          config: {
            case: 'generateBoolConfig',
            value: create(GenerateBoolSchema, {}),
          },
        }),
      })
    ),
    byte: convertJobMappingTransformerToForm(
      create(JobMappingTransformerSchema, {
        config: create(TransformerConfigSchema, {
          config: {
            case: 'passthroughConfig',
            value: create(PassthroughSchema, {}),
          },
        }),
      })
    ),
    n: convertJobMappingTransformerToForm(
      create(JobMappingTransformerSchema, {
        config: create(TransformerConfigSchema, {
          config: {
            case: 'passthroughConfig',
            value: create(PassthroughSchema, {}),
          },
        }),
      })
    ),
    s: convertJobMappingTransformerToForm(
      create(JobMappingTransformerSchema, {
        config: create(TransformerConfigSchema, {
          config: {
            case: 'generateStringConfig',
            value: create(GenerateStringSchema, {
              min: BigInt(1),
              max: BigInt(100),
            }),
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
        map[subset.schema] = create(PostgresSourceSchemaOptionSchema, {
          schema: subset.schema,
          tables: [],
        });
      }
      map[subset.schema].tables.push(
        create(PostgresSourceTableOptionSchema, {
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
        map[subset.schema] = create(MysqlSourceSchemaOptionSchema, {
          schema: subset.schema,
          tables: [],
        });
      }
      map[subset.schema].tables.push(
        create(MysqlSourceTableOptionSchema, {
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
        map[subset.schema] = create(MssqlSourceSchemaOptionSchema, {
          schema: subset.schema,
          tables: [],
        });
      }
      map[subset.schema].tables.push(
        create(MssqlSourceTableOptionSchema, {
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
  return subsets.map((ss) =>
    create(DynamoDBSourceTableOptionSchema, {
      table: ss.table,
      whereClause: ss.whereClause,
    })
  );
}

function toDynamoDbSourceUnmappedOptions(
  unmappedTransformConfig: DynamoDBSourceUnmappedTransformConfigFormValues
): DynamoDBSourceUnmappedTransformConfig {
  return create(DynamoDBSourceUnmappedTransformConfigSchema, {
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
        create(JobMappingTransformerSchema, {
          config: create(TransformerConfigSchema, {
            config: {
              case: 'generateBoolConfig',
              value: create(GenerateBoolSchema, {}),
            },
          }),
        })
    ),
    byte: convertJobMappingTransformerToForm(
      ut.b ||
        create(JobMappingTransformerSchema, {
          config: create(TransformerConfigSchema, {
            config: {
              case: 'passthroughConfig',
              value: create(PassthroughSchema, {}),
            },
          }),
        })
    ),
    n: convertJobMappingTransformerToForm(
      ut.n ||
        create(JobMappingTransformerSchema, {
          config: create(TransformerConfigSchema, {
            config: {
              case: 'passthroughConfig',
              value: create(PassthroughSchema, {}),
            },
          }),
        })
    ),
    s: convertJobMappingTransformerToForm(
      ut.s ||
        create(JobMappingTransformerSchema, {
          config: create(TransformerConfigSchema, {
            config: {
              case: 'generateStringConfig',
              value: create(GenerateStringSchema, {
                min: BigInt(1),
                max: BigInt(100),
              }),
            },
          }),
        })
    ),
  };
}

export function toActivityOptions(
  values: ActivityOptionsFormValues
): ActivityOptions {
  return create(ActivityOptionsSchema, {
    startToCloseTimeout:
      values.startToCloseTimeout !== undefined && values.startToCloseTimeout > 0
        ? convertMinutesToNanoseconds(values.startToCloseTimeout)
        : undefined,
    scheduleToCloseTimeout:
      values.scheduleToCloseTimeout !== undefined &&
      values.scheduleToCloseTimeout > 0
        ? convertMinutesToNanoseconds(values.scheduleToCloseTimeout)
        : undefined,
    retryPolicy: values.retryPolicy
      ? create(RetryPolicySchema, {
          maximumAttempts: values.retryPolicy.maximumAttempts,
        })
      : undefined,
  });
}

export function toJobSourceSqlSubsetSchemas(
  values: SubsetFormValues,
  dbType: ValidSubsetConnectionType | null
): JobSourceSqlSubetSchemas {
  switch (dbType) {
    case 'mysqlConfig': {
      return create(JobSourceSqlSubetSchemasSchema, {
        schemas: {
          case: 'mysqlSubset',
          value: create(MysqlSourceSchemaSubsetSchema, {
            mysqlSchemas: toMysqlSourceSchemaOptions(values.subsets),
          }),
        },
      });
    }
    case 'pgConfig': {
      return create(JobSourceSqlSubetSchemasSchema, {
        schemas: {
          case: 'postgresSubset',
          value: create(PostgresSourceSchemaSubsetSchema, {
            postgresSchemas: toPostgresSourceSchemaOptions(values.subsets),
          }),
        },
      });
    }
    case 'dynamodbConfig': {
      return create(JobSourceSqlSubetSchemasSchema, {
        schemas: {
          case: 'dynamodbSubset',
          value: create(DynamoDBSourceSchemaSubsetSchema, {
            tables: toDynamoDbSourceTableOptions(values.subsets),
          }),
        },
      });
    }
    case 'mssqlConfig': {
      return create(JobSourceSqlSubetSchemasSchema, {
        schemas: {
          case: 'mssqlSubset',
          value: create(MssqlSourceSchemaSubsetSchema, {
            mssqlSchemas: toMssqlSourceSchemaOptions(values.subsets),
          }),
        },
      });
    }
    default: {
      return create(JobSourceSqlSubetSchemasSchema, {});
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
            ? getDestinationFormValuesOrDefaultFromDestination(
                job.destinations[0]
              )
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
            ? getDestinationFormValuesOrDefaultFromDestination(
                job.destinations[0]
              )
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
          getDestinationFormValuesOrDefaultFromDestination(dest)
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
          getDestinationFormValuesOrDefaultFromDestination(dest)
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
            columnRemovalStrategy: toColumnRemovalStrategy(
              job.source.options.config.value.columnRemovalStrategy
            ),
          },
        },
        destinations: job.destinations.map((dest) =>
          getDestinationFormValuesOrDefaultFromDestination(dest)
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
            columnRemovalStrategy: toColumnRemovalStrategy(
              job.source.options.config.value.columnRemovalStrategy
            ),
          },
        },
        destinations: job.destinations.map((dest) =>
          getDestinationFormValuesOrDefaultFromDestination(dest)
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
            columnRemovalStrategy: toColumnRemovalStrategy(
              job.source.options.config.value.columnRemovalStrategy
            ),
          },
        },
        destinations: job.destinations.map((dest) =>
          getDestinationFormValuesOrDefaultFromDestination(dest)
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
              : convertJobMappingTransformerToForm(
                  create(JobMappingTransformerSchema, {})
                ),
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
              : convertJobMappingTransformerToForm(
                  create(JobMappingTransformerSchema, {})
                ),
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
              : convertJobMappingTransformerToForm(
                  create(JobMappingTransformerSchema, {})
                ),
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

const DEFAULT_MAX_FLIGHT = 10;
const DEFAULT_BATCH_COUNT = 100;
const DEFAULT_BATCH_PERIOD = '5s';

export function getDefaultPostgresDestinationFormValueOptions(): PostgresDbDestinationOptionsFormValues {
  return {
    initTableSchema: false,
    conflictStrategy: {
      onConflictDoNothing: false,
      onConflictDoUpdate: false,
    },
    skipForeignKeyViolations: false,
    truncateBeforeInsert: false,
    truncateCascade: false,
    batch: {
      count: DEFAULT_BATCH_COUNT,
      period: DEFAULT_BATCH_PERIOD,
    },
    maxInFlight: DEFAULT_MAX_FLIGHT,
  };
}

export function getDefaultMysqlDestinationFormValueOptions(): MysqlDbDestinationOptionsFormValues {
  return {
    initTableSchema: false,
    conflictStrategy: {
      onConflictDoNothing: false,
      onConflictDoUpdate: false,
    },
    skipForeignKeyViolations: false,
    truncateBeforeInsert: false,
    batch: {
      count: DEFAULT_BATCH_COUNT,
      period: DEFAULT_BATCH_PERIOD,
    },
    maxInFlight: DEFAULT_MAX_FLIGHT,
  };
}

export function getDefaultMssqlDestinationFormValueOptions(): MssqlDbDestinationOptionsFormValues {
  return {
    initTableSchema: false,
    onConflictDoNothing: false,
    skipForeignKeyViolations: false,
    truncateBeforeInsert: false,
    batch: {
      count: DEFAULT_BATCH_COUNT,
      period: DEFAULT_BATCH_PERIOD,
    },
    maxInFlight: DEFAULT_MAX_FLIGHT,
  };
}

export function getDefaultAwsS3DestinationFormValueOptions(): AwsS3DestinationOptionsFormValues {
  return {
    storageClass: AwsS3DestinationConnectionOptions_StorageClass.STANDARD,
    timeout: '5s',
    batch: {
      count: DEFAULT_BATCH_COUNT,
      period: DEFAULT_BATCH_PERIOD,
    },
    maxInFlight: DEFAULT_MAX_FLIGHT,
  };
}

export function getDefaultDestinationFormValueOptionsFromConnectionCase(
  destCase: ConnectionConfigCase | null | undefined,
  getUniqueTables: () => Set<string>
): DestinationOptionsFormValues {
  switch (destCase) {
    case 'pgConfig': {
      return {
        postgres: getDefaultPostgresDestinationFormValueOptions(),
      };
    }
    case 'mysqlConfig': {
      return {
        mysql: getDefaultMysqlDestinationFormValueOptions(),
      };
    }
    case 'mssqlConfig': {
      return {
        mssql: getDefaultMssqlDestinationFormValueOptions(),
      };
    }
    case 'awsS3Config': {
      return {
        awss3: getDefaultAwsS3DestinationFormValueOptions(),
      };
    }
    case 'dynamodbConfig': {
      return {
        dynamodb: {
          tableMappings: Array.from(getUniqueTables()).map((table) => ({
            sourceTable: table,
            destinationTable: '',
          })),
        },
      };
    }
    case 'mongoConfig': {
      return {};
    }
    case 'gcpCloudstorageConfig': {
      return {};
    }
    case 'localDirConfig': {
      return {};
    }
    case 'openaiConfig': {
      return {};
    }
    default: {
      return {};
    }
  }
}

// Based on the job destiation type, returns the form values, or their default equivalent
export function getDestinationFormValuesOrDefaultFromDestination(
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
            conflictStrategy: {
              onConflictDoNothing:
                d.options.config.value.onConflict?.strategy.case === 'nothing'
                  ? true
                  : false,
              onConflictDoUpdate:
                d.options.config.value.onConflict?.strategy.case === 'update'
                  ? true
                  : false,
            },
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
            conflictStrategy: {
              onConflictDoNothing:
                d.options.config.value.onConflict?.strategy.case === 'nothing'
                  ? true
                  : false,
              onConflictDoUpdate:
                d.options.config.value.onConflict?.strategy.case === 'update'
                  ? true
                  : false,
            },
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
  piidetect: { connect: string; schema: string };
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
    piidetect: {
      connect: `${sessionId}-new-job-pii-detect-connect`,
      schema: `${sessionId}-new-job-pii-detect-schema`,
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
  formMappings: JobMappingFormValues[],
  accountId: string,
  virtualForeignKeys: VirtualForeignConstraintFormValues[],
  validate: (
    req: ValidateJobMappingsRequest
  ) => Promise<ValidateJobMappingsResponse>,
  jobSource?: JobSource
): Promise<ValidateJobMappingsResponse> {
  const body = create(ValidateJobMappingsRequestSchema, {
    accountId,
    mappings: formMappings.map((m) => {
      return create(JobMappingSchema, {
        schema: m.schema,
        table: m.table,
        column: m.column,
        transformer: m.transformer.config.case
          ? convertJobMappingTransformerFormToJobMappingTransformer(
              m.transformer
            )
          : create(JobMappingTransformerSchema, {
              config: create(TransformerConfigSchema, {
                config: {
                  case: 'passthroughConfig',
                  value: {},
                },
              }),
            }),
      });
    }),
    virtualForeignKeys: virtualForeignKeys.map((v) => {
      return create(VirtualForeignConstraintSchema, {
        schema: v.schema,
        table: v.table,
        columns: v.columns,
        foreignKey: create(VirtualForeignKeySchema, {
          schema: v.foreignKey.schema,
          table: v.foreignKey.table,
          columns: v.foreignKey.columns,
        }),
      });
    }),
    connectionId: getConnectionIdFromSource(jobSource),
    jobSource: jobSource,
  });

  return validate(body);
}
