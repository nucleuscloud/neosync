'use client';
import SourceOptionsForm from '@/components/jobs/Form/SourceOptionsForm';
import NosqlTable from '@/components/jobs/NosqlTable/NosqlTable';
import { OnTableMappingUpdateRequest } from '@/components/jobs/NosqlTable/TableMappings/Columns';
import {
  SchemaTable,
  getAllFormErrors,
} from '@/components/jobs/SchemaTable/SchemaTable';
import { getSchemaConstraintHandler } from '@/components/jobs/SchemaTable/schema-constraint-handler';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useToast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
import {
  DataSyncSourceFormValues,
  DestinationOptionFormValues,
  VirtualForeignConstraintFormValues,
  convertJobMappingTransformerFormToJobMappingTransformer,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import { PartialMessage } from '@bufbuild/protobuf';
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  Connection,
  DynamoDBSourceConnectionOptions,
  GetConnectionResponse,
  GetConnectionSchemaMapRequest,
  GetConnectionSchemaMapResponse,
  GetConnectionSchemaMapsResponse,
  GetConnectionSchemaResponse,
  Job,
  JobMapping,
  JobMappingTransformer,
  JobSource,
  JobSourceOptions,
  MongoDBSourceConnectionOptions,
  MysqlSourceConnectionOptions,
  PostgresSourceConnectionOptions,
  ValidateJobMappingsResponse,
  VirtualForeignConstraint,
  VirtualForeignKey,
} from '@neosync/sdk';
import {
  getConnection,
  getConnectionSchemaMap,
  getConnectionSchemaMaps,
  getConnectionTableConstraints,
  getConnections,
  getJob,
  updateJobDestinationConnection,
  updateJobSourceConnection,
  validateJobMappings,
} from '@neosync/sdk/connectquery';
import { useQueryClient } from '@tanstack/react-query';
import { ReactElement, useCallback, useEffect, useMemo, useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import { validateJobMapping } from '../../../util';
import SchemaPageSkeleton from './SchemaPageSkeleton';
import {
  getDestinationDetailsRecord,
  getDynamoDbDestinations,
  getOnSelectedTableToggle,
  isDynamoDBConnection,
  isNosqlSource,
  shouldShowDestinationTableMappings,
} from './util';

interface Props {
  jobId: string;
}

function getConnectionIdFromSource(
  js: JobSource | undefined
): string | undefined {
  if (
    js?.options?.config.case === 'postgres' ||
    js?.options?.config.case === 'mysql' ||
    js?.options?.config.case === 'awsS3' ||
    js?.options?.config.case === 'mongodb' ||
    js?.options?.config.case === 'dynamodb'
  ) {
    return js.options.config.value.connectionId;
  }
  return undefined;
}

export default function DataSyncConnectionCard({ jobId }: Props): ReactElement {
  const { toast } = useToast();
  const { account } = useAccount();
  const {
    data,
    refetch: mutate,
    isLoading: isJobDataLoading,
  } = useQuery(getJob, { id: jobId }, { enabled: !!jobId });
  const sourceConnectionId = getConnectionIdFromSource(data?.job?.source);

  const {
    data: connectionSchemaDataMap,
    isLoading: isSchemaDataMapLoading,
    isFetching: isSchemaMapValidating,
  } = useQuery(
    getConnectionSchemaMap,
    { connectionId: sourceConnectionId },
    { enabled: !!sourceConnectionId }
  );
  const { mutateAsync: getConnectionSchemaMapAsync } = useMutation(
    getConnectionSchemaMap
  );

  const { data: destinationConnectionSchemaMapsResp } = useQuery(
    getConnectionSchemaMaps,
    {
      requests: data?.job?.destinations.map(
        (dest): PartialMessage<GetConnectionSchemaMapRequest> => ({
          connectionId: dest.connectionId,
        })
      ),
    },
    {
      enabled:
        (data?.job?.destinations.length ?? 0) > 0 &&
        data?.job?.source?.options?.config.case === 'dynamodb',
    }
  );

  const { isLoading: isConnectionsLoading, data: connectionsData } = useQuery(
    getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const connections = connectionsData?.connections ?? [];
  const connectionsRecord = connections.reduce(
    (record, conn) => {
      record[conn.id] = conn;
      return record;
    },
    {} as Record<string, Connection>
  );

  const { mutateAsync: updateJobSrcConnection } = useMutation(
    updateJobSourceConnection
  );
  const { mutateAsync: updateJobDestConnection } = useMutation(
    updateJobDestinationConnection
  );

  const queryclient = useQueryClient();

  const [validateMappingsResponse, setValidateMappingsResponse] = useState<
    ValidateJobMappingsResponse | undefined
  >();

  const [isValidatingMappings, setIsValidatingMappings] = useState(false);

  const form = useForm<DataSyncSourceFormValues>({
    resolver: yupResolver<DataSyncSourceFormValues>(DataSyncSourceFormValues),
    values: getJobSource(data?.job, connectionSchemaDataMap?.schemaMap),
    context: { accountId: account?.id },
  });

  const { data: tableConstraints, isFetching: isTableConstraintsValidating } =
    useQuery(
      getConnectionTableConstraints,
      { connectionId: sourceConnectionId },
      { enabled: !!sourceConnectionId }
    );
  const { mutateAsync: getConnectionAsync } = useMutation(getConnection);

  const {
    append: appendVfk,
    remove: removeVfk,
    fields: formVirtualForeignKeys,
  } = useFieldArray({
    control: form.control,
    name: 'virtualForeignKeys',
  });

  const schemaConstraintHandler = useMemo(() => {
    const virtualForeignKeys = Array.from(data?.job?.virtualForeignKeys ?? []);
    formVirtualForeignKeys?.forEach((v) => {
      virtualForeignKeys.push(
        new VirtualForeignConstraint({
          schema: v.schema,
          table: v.table,
          columns: v.columns,
          foreignKey: new VirtualForeignKey({
            schema: v.foreignKey.schema,
            table: v.foreignKey.table,
            columns: v.foreignKey.columns,
          }),
        })
      );
    });

    return getSchemaConstraintHandler(
      connectionSchemaDataMap?.schemaMap ?? {},
      tableConstraints?.primaryKeyConstraints ?? {},
      tableConstraints?.foreignKeyConstraints ?? {},
      tableConstraints?.uniqueConstraints ?? {},
      virtualForeignKeys
    );
  }, [
    isSchemaMapValidating,
    isTableConstraintsValidating,
    isJobDataLoading,
    formVirtualForeignKeys, // this is kinda dangerous
  ]);
  const [selectedTables, setSelectedTables] = useState<Set<string>>(new Set());

  const {
    append,
    remove,
    update,
    fields: formMappings,
  } = useFieldArray({
    control: form.control,
    name: 'mappings',
  });

  // const {
  //   append: appendDestOpts,
  //   remove: removeDestOpts,
  //   update: updateDestOpts,
  //   fields: destOptsValues,
  // } = useFieldArray<DataSyncSourceFormValues>({
  //   control: form.control,
  //   name: 'destinationOptions',
  // });

  useEffect(() => {
    if (isJobDataLoading || isSchemaDataMapLoading || selectedTables.size > 0) {
      return;
    }
    const js = getJobSource(data?.job, connectionSchemaDataMap?.schemaMap);
    setSelectedTables(
      new Set(
        js.mappings.map((mapping) => `${mapping.schema}.${mapping.table}`)
      )
    );
  }, [isJobDataLoading, isSchemaDataMapLoading]);

  const { mutateAsync: validateJobMappingsAsync } =
    useMutation(validateJobMappings);

  async function onSourceChange(value: string): Promise<void> {
    try {
      const newValues = await getUpdatedValues(
        value,
        form.getValues(),
        async (id) => {
          const resp = await getConnectionAsync({ id });
          queryclient.setQueryData(
            createConnectQueryKey(getConnection, { id }),
            resp
          );
          return resp;
        },
        async (id) => {
          const resp = await getConnectionSchemaMapAsync({ connectionId: id });
          queryclient.setQueryData(
            createConnectQueryKey(getConnectionSchemaMap, { connectionId: id }),
            resp
          );
          return resp;
        }
      );
      form.reset(newValues);
    } catch (err) {
      form.reset({ ...form.getValues(), mappings: [], sourceId: value });
      toast({
        title: 'Unable to get connection schema',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  async function onSubmit(values: DataSyncSourceFormValues) {
    const connection = connections.find((c) => c.id === values.sourceId);
    const job = data?.job;
    if (!job || !connection) {
      return;
    }
    try {
      await updateJobSrcConnection({
        id: job.id,
        mappings: values.mappings.map((m) => {
          return new JobMapping({
            schema: m.schema,
            table: m.table,
            column: m.column,
            transformer:
              convertJobMappingTransformerFormToJobMappingTransformer(
                m.transformer
              ),
          });
        }),
        virtualForeignKeys:
          values.virtualForeignKeys?.map((v) => {
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
          }) || [],
        source: new JobSource({
          options: toJobSourceOptions(values, job, connection, values.sourceId),
        }),
      });
      toast({
        title: 'Successfully updated job source connection!',
        variant: 'success',
      });
      // hold off on mutating until after we update the job dest connections for dynamo conns
      if (connection.connectionConfig?.config.case !== 'dynamodbConfig') {
        mutate();
      }
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to update job source connection',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
      return;
    }
    if (connection.connectionConfig?.config.case !== 'dynamodbConfig') {
      return;
    }
    try {
      const destIdToConnId = new Map(
        data?.job?.destinations.map((d) => [d.id, d.connectionId])
      );
      await Promise.all(
        values.destinationOptions.map(async (destOpts) => {
          if (!destOpts.dynamoDb) {
            return;
          }
          return updateJobDestConnection({
            destinationId: destOpts.destinationId,
            jobId: data?.job?.id,
            connectionId: destIdToConnId.get(destOpts.destinationId),
            options: {
              config: {
                case: 'dynamodbOptions',
                value: {
                  tableMappings: destOpts.dynamoDb.tableMappings ?? [],
                },
              },
            },
          });
        })
      );
      toast({
        title: 'Successfully updated job destination connection(s)',
        variant: 'success',
      });
      mutate();
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to update one or all job destination connections',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  async function validateMappings() {
    try {
      setIsValidatingMappings(true);
      const res = await validateJobMapping(
        sourceConnectionId || '',
        formMappings,
        account?.id || '',
        formVirtualForeignKeys,
        validateJobMappingsAsync
      );
      setValidateMappingsResponse(res);
    } catch (error) {
      console.error('Failed to validate job mappings:', error);
      toast({
        title: 'Unable to validate job mappings',
        variant: 'destructive',
      });
    } finally {
      setIsValidatingMappings(false);
    }
  }

  async function validateVirtualForeignKeys(
    vfks: VirtualForeignConstraintFormValues[]
  ) {
    try {
      setIsValidatingMappings(true);
      const res = await validateJobMapping(
        sourceConnectionId || '',
        formMappings,
        account?.id || '',
        vfks,
        validateJobMappingsAsync
      );
      setValidateMappingsResponse(res);
    } catch (error) {
      console.error('Failed to validate virtual foreign keys:', error);
      toast({
        title: 'Unable to validate virtual foreign keys',
        variant: 'destructive',
      });
    } finally {
      setIsValidatingMappings(false);
    }
  }

  const onSelectedTableToggle = getOnSelectedTableToggle(
    connectionSchemaDataMap?.schemaMap ?? {},
    selectedTables,
    setSelectedTables,
    formMappings,
    remove,
    append
  );

  useEffect(() => {
    if (!account?.id || !sourceConnectionId) {
      return;
    }
    const validateJobMappings = async () => {
      await validateMappings();
    };
    validateJobMappings();
  }, [selectedTables, account?.id, sourceConnectionId]);

  async function addVirtualForeignKey(vfk: VirtualForeignConstraintFormValues) {
    appendVfk(vfk);
    const vfks = [vfk, ...(formVirtualForeignKeys || [])];
    await validateVirtualForeignKeys(vfks);
  }

  async function removeVirtualForeignKey(index: number) {
    const newVfks: VirtualForeignConstraintFormValues[] = [];
    formVirtualForeignKeys?.forEach((vfk, idx) => {
      if (idx != index) {
        newVfks.push(vfk);
      }
    });
    removeVfk(index);
    await validateVirtualForeignKeys(newVfks);
  }

  const onDestinationTableMappingUpdate = useCallback(
    (req: OnTableMappingUpdateRequest) => {
      const destOpts = form.getValues('destinationOptions');
      const destOpt = destOpts.find(
        (d) => d.destinationId === req.destinationId
      );
      const tm = destOpt?.dynamoDb?.tableMappings.find(
        (tm) => tm.sourceTable === req.souceName
      );
      if (tm) {
        tm.destinationTable = req.tableName;
        form.setValue('destinationOptions', destOpts);
      }
      return;
    },
    []
  );

  if (isConnectionsLoading || isSchemaDataMapLoading || isJobDataLoading) {
    return <SchemaPageSkeleton />;
  }

  const source = connectionsRecord[sourceConnectionId ?? ''] as
    | Connection
    | undefined;

  const dynamoDBDestinations = getDynamoDbDestinations(
    data?.job?.destinations ?? []
  );

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        <div className="space-y-8">
          <FormField
            control={form.control}
            name="sourceId"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Source</FormLabel>
                <FormDescription>
                  The location of the source data set.
                </FormDescription>
                <FormControl>
                  <Select
                    value={field.value}
                    onValueChange={async (value) => {
                      if (!value) {
                        return;
                      }
                      field.onChange(value);
                      await onSourceChange(value);
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={source?.name} />
                    </SelectTrigger>
                    <SelectContent>
                      {connections
                        .filter(
                          (c) =>
                            !data?.job?.destinations
                              .map((d) => d.connectionId)
                              ?.includes(c.id) &&
                            c.connectionConfig?.config.case !== 'awsS3Config' &&
                            c.connectionConfig?.config.case !==
                              'openaiConfig' &&
                            c.connectionConfig?.config.case !==
                              'gcpCloudstorageConfig'
                        )
                        .map((connection) => (
                          <SelectItem
                            className="cursor-pointer ml-2"
                            key={connection.id}
                            value={connection.id}
                          >
                            {connection.name}
                          </SelectItem>
                        ))}
                    </SelectContent>
                  </Select>
                </FormControl>

                <FormMessage />
              </FormItem>
            )}
          />
          <SourceOptionsForm
            connection={connections.find(
              (c) => c.id === form.getValues().sourceId
            )}
          />

          {isNosqlSource(source ?? new Connection()) && (
            <NosqlTable
              data={formMappings}
              destinationOptions={form.watch('destinationOptions')}
              schema={connectionSchemaDataMap?.schemaMap ?? {}}
              isSchemaDataReloading={isSchemaMapValidating}
              isJobMappingsValidating={isValidatingMappings}
              formErrors={getAllFormErrors(
                form.formState.errors,
                formMappings,
                validateMappingsResponse
              )}
              onValidate={validateMappings}
              constraintHandler={schemaConstraintHandler}
              onRemoveMappings={(values) => {
                const valueSet = new Set(
                  values.map((v) => `${v.schema}.${v.table}.${v.column}`)
                );
                const toRemove: number[] = [];
                formMappings.forEach((mapping, idx) => {
                  if (
                    valueSet.has(
                      `${mapping.schema}.${mapping.table}.${mapping.column}`
                    )
                  ) {
                    toRemove.push(idx);
                  }
                });
                if (toRemove.length > 0) {
                  remove(toRemove);
                }

                if (!source || isDynamoDBConnection(source)) {
                  return;
                }

                const toRemoveSet = new Set(toRemove);
                const remainingTables = formMappings
                  .filter((_, idx) => !toRemoveSet.has(idx))
                  .map((fm) => fm.table);

                // Check and update destinationOptions if needed
                const destOpts = form.getValues('destinationOptions');
                const updatedDestOpts = destOpts
                  .map((opt) => {
                    if (opt.dynamoDb) {
                      const updatedTableMappings =
                        opt.dynamoDb.tableMappings.filter((tm) => {
                          // Check if any columns remain for the table
                          const tableColumnsExist = remainingTables.some(
                            (table) => table === tm.sourceTable
                          );
                          return tableColumnsExist;
                        });

                      return {
                        ...opt,
                        dynamoDb: {
                          ...opt.dynamoDb,
                          tableMappings: updatedTableMappings,
                        },
                      };
                    }
                    return opt;
                  })
                  .filter(
                    (opt) => (opt.dynamoDb?.tableMappings.length ?? 0) > 0
                  );
                form.setValue('destinationOptions', updatedDestOpts);
              }}
              onEditMappings={(values) => {
                const valuesMap = new Map(
                  values.map((v) => [
                    `${v.schema}.${v.table}.${v.column}`,
                    v.transformer,
                  ])
                );
                formMappings.forEach((fm, idx) => {
                  const fmKey = `${fm.schema}.${fm.table}.${fm.column}`;
                  const fmTrans = valuesMap.get(fmKey);
                  if (fmTrans) {
                    update(idx, {
                      ...fm,
                      transformer: fmTrans,
                    });
                  }
                });
              }}
              onAddMappings={(values) => {
                append(
                  values.map((v) => {
                    const [schema, table] = v.collection.split('.');
                    return {
                      schema,
                      table,
                      column: v.key,
                      transformer: v.transformer,
                    };
                  })
                );
                const uniqueCollections = Array.from(
                  new Set(values.map((v) => v.collection))
                );

                const destOpts = form.getValues('destinationOptions');
                const existing = new Map(
                  destOpts.map((d) => [d.destinationId, d])
                );
                const updated = dynamoDBDestinations.map(
                  (dest): DestinationOptionFormValues => {
                    const opt = existing.get(dest.id);
                    if (opt) {
                      const sourceSet = new Set(
                        opt.dynamoDb?.tableMappings.map(
                          (mapping) => mapping.sourceTable
                        ) ?? []
                      );

                      // Add missing uniqueCollections to the existing tableMappings
                      const updatedTableMappings = [
                        ...(opt.dynamoDb?.tableMappings ?? []),
                        ...uniqueCollections
                          .map((c) => {
                            const [, table] = c.split('.');
                            return {
                              sourceTable: table,
                              destinationTable: '',
                            };
                          })
                          .filter(
                            (mapping) => !sourceSet.has(mapping.sourceTable)
                          ),
                      ];

                      return {
                        ...opt,
                        dynamoDb: {
                          ...opt.dynamoDb,
                          tableMappings: updatedTableMappings,
                        },
                      };
                    }

                    return {
                      destinationId: dest.id,
                      dynamoDb: {
                        tableMappings: uniqueCollections.map((c) => {
                          const [, table] = c.split('.');
                          return {
                            sourceTable: table,
                            destinationTable: 'todo',
                          };
                        }),
                      },
                    };
                  }
                );

                form.setValue('destinationOptions', updated);
              }}
              destinationDetailsRecord={getDestinationDetailsRecord(
                dynamoDBDestinations,
                connectionsRecord,
                destinationConnectionSchemaMapsResp ??
                  new GetConnectionSchemaMapsResponse()
              )}
              onDestinationTableMappingUpdate={onDestinationTableMappingUpdate}
              showDestinationTableMappings={shouldShowDestinationTableMappings(
                source ?? new Connection(),
                dynamoDBDestinations.length > 0
              )}
            />
          )}

          {!isNosqlSource(source ?? new Connection()) && (
            <SchemaTable
              data={formMappings}
              virtualForeignKeys={formVirtualForeignKeys}
              jobType="sync"
              constraintHandler={schemaConstraintHandler}
              schema={connectionSchemaDataMap?.schemaMap ?? {}}
              isSchemaDataReloading={isSchemaMapValidating}
              selectedTables={selectedTables}
              onSelectedTableToggle={onSelectedTableToggle}
              formErrors={getAllFormErrors(
                form.formState.errors,
                formMappings,
                validateMappingsResponse
              )}
              isJobMappingsValidating={isValidatingMappings}
              onValidate={validateMappings}
              addVirtualForeignKey={addVirtualForeignKey}
              removeVirtualForeignKey={removeVirtualForeignKey}
            />
          )}
          <div className="flex flex-row items-center justify-end w-full mt-4">
            <Button type="submit">Update</Button>
          </div>
        </div>
      </form>
    </Form>
  );
}

function toJobSourceOptions(
  values: DataSyncSourceFormValues,
  job: Job,
  connection: Connection,
  newSourceId: string
): JobSourceOptions {
  switch (connection.connectionConfig?.config.case) {
    case 'pgConfig':
      return new JobSourceOptions({
        config: {
          case: 'postgres',
          value: new PostgresSourceConnectionOptions({
            ...getExistingPostgresSourceConnectionOptions(job),
            connectionId: newSourceId,
            haltOnNewColumnAddition:
              values.sourceOptions.haltOnNewColumnAddition,
          }),
        },
      });
    case 'mysqlConfig':
      return new JobSourceOptions({
        config: {
          case: 'mysql',
          value: new MysqlSourceConnectionOptions({
            ...getExistingMysqlSourceConnectionOptions(job),
            connectionId: newSourceId,
            haltOnNewColumnAddition:
              values.sourceOptions.haltOnNewColumnAddition,
          }),
        },
      });
    case 'mongoConfig':
      return new JobSourceOptions({
        config: {
          case: 'mongodb',
          value: new MongoDBSourceConnectionOptions({
            ...getExistingMongoSourceConnectionOptions(job),
            connectionId: newSourceId,
          }),
        },
      });
    case 'dynamodbConfig': {
      return new JobSourceOptions({
        config: {
          case: 'dynamodb',
          value: new DynamoDBSourceConnectionOptions({
            ...getExistingDynamoDBSourceConnectionOptions(job),
            connectionId: newSourceId,
          }),
        },
      });
    }
    default:
      throw new Error('unsupported connection type');
  }
}

function getExistingPostgresSourceConnectionOptions(
  job: Job
): PostgresSourceConnectionOptions | undefined {
  return job.source?.options?.config.case === 'postgres'
    ? job.source.options.config.value
    : undefined;
}

function getExistingMysqlSourceConnectionOptions(
  job: Job
): MysqlSourceConnectionOptions | undefined {
  return job.source?.options?.config.case === 'mysql'
    ? job.source.options.config.value
    : undefined;
}

function getExistingMongoSourceConnectionOptions(
  job: Job
): MongoDBSourceConnectionOptions | undefined {
  return job.source?.options?.config.case === 'mongodb'
    ? job.source.options.config.value
    : undefined;
}

function getExistingDynamoDBSourceConnectionOptions(
  job: Job
): DynamoDBSourceConnectionOptions | undefined {
  return job.source?.options?.config.case === 'dynamodb'
    ? job.source.options.config.value
    : undefined;
}

function getJobSource(
  job?: Job,
  connSchemaMap?: Record<string, GetConnectionSchemaResponse>
): DataSyncSourceFormValues {
  if (!job || !connSchemaMap) {
    return {
      sourceId: '',
      sourceOptions: {
        haltOnNewColumnAddition: false,
      },
      mappings: [],
      virtualForeignKeys: [],
      connectionId: '',
      destinationOptions: [],
    };
  }

  const mapData: Record<string, Set<string>> = {};

  const mappings = (job.mappings ?? []).map((mapping) => {
    const tkey = `${mapping.schema}.${mapping.table}`;
    const uniqcols = mapData[tkey];
    if (uniqcols) {
      uniqcols.add(mapping.column);
    } else {
      mapData[tkey] = new Set([mapping.column]);
    }

    return {
      ...mapping,
      transformer: mapping.transformer
        ? convertJobMappingTransformerToForm(mapping.transformer)
        : convertJobMappingTransformerToForm(new JobMappingTransformer()),
    };
  });

  const virtualForeignKeys = (job.virtualForeignKeys ?? []).map((vfk) => {
    return {
      ...vfk,
      foreignKey: {
        schema: vfk.foreignKey?.schema || '',
        table: vfk.foreignKey?.table || '',
        columns: vfk.foreignKey?.columns || [],
      },
    };
  });

  if (
    job.source?.options?.config.case === 'postgres' ||
    job.source?.options?.config.case === 'mysql'
  ) {
    Object.entries(mapData).forEach(([key, currcols]) => {
      const dbcols = connSchemaMap[key];
      if (!dbcols) {
        return;
      }
      dbcols.schemas.forEach((dbcol) => {
        if (!currcols.has(dbcol.column)) {
          mappings.push({
            schema: dbcol.schema,
            table: dbcol.table,
            column: dbcol.column,
            transformer: convertJobMappingTransformerToForm(
              new JobMappingTransformer()
            ),
          });
        }
      });
    });
  }

  const values = {
    sourceOptions: {},
    mappings: mappings || [],
    virtualForeignKeys: virtualForeignKeys || [],
  };
  const yupValidationValues = {
    ...values,
    sourceId: getConnectionIdFromSource(job.source) || '',
    mappings,
    connectionId: getConnectionIdFromSource(job.source) || '',
    destinationOptions: [],
  };

  switch (job?.source?.options?.config.case) {
    case 'postgres':
      return {
        ...yupValidationValues,
        sourceId: getConnectionIdFromSource(job.source) || '',
        sourceOptions: {
          haltOnNewColumnAddition:
            job?.source?.options?.config.value.haltOnNewColumnAddition,
        },
      };
    case 'mysql':
      return {
        ...yupValidationValues,
        sourceId: getConnectionIdFromSource(job.source) || '',
        sourceOptions: {
          haltOnNewColumnAddition:
            job?.source?.options?.config.value.haltOnNewColumnAddition,
        },
      };
    case 'mongodb':
      return {
        ...yupValidationValues,
        sourceId: getConnectionIdFromSource(job.source) || '',
        sourceOptions: {},
      };
    case 'dynamodb': {
      const destOpts: DestinationOptionFormValues[] = [];
      job.destinations.forEach((d) => {
        if (d.options?.config.case !== 'dynamodbOptions') {
          return;
        }
        destOpts.push({
          destinationId: d.id,
          dynamoDb: {
            tableMappings: d.options.config.value.tableMappings,
          },
        });
      });
      return {
        ...yupValidationValues,
        sourceId: getConnectionIdFromSource(job.source) || '',
        sourceOptions: {},
        destinationOptions: destOpts,
      };
    }
    default:
      return yupValidationValues;
  }
}

async function getUpdatedValues(
  connectionId: string,
  originalValues: DataSyncSourceFormValues,
  getConnectionById: (id: string) => Promise<GetConnectionResponse>,
  getConnectionSchemaMapAsync: (
    id: string
  ) => Promise<GetConnectionSchemaMapResponse>
): Promise<DataSyncSourceFormValues> {
  const [schemaRes, connRes] = await Promise.all([
    getConnectionSchemaMapAsync(connectionId),
    getConnectionById(connectionId),
  ]);

  if (!schemaRes || !connRes) {
    return originalValues;
  }

  const sameKeys = new Set(
    Object.values(schemaRes.schemaMap).flatMap((dbcols) =>
      dbcols.schemas.map(
        (dbcol) => `${dbcol.schema}.${dbcol.table}.${dbcol.column}`
      )
    )
  );

  const mappings = originalValues.mappings.filter((mapping) =>
    sameKeys.has(`${mapping.schema}.${mapping.table}.${mapping.column}`)
  );

  const values = {
    sourceId: connectionId || '',
    sourceOptions: {},
    // destinationIds: originalValues.destinationIds,
    mappings,
    connectionId: connectionId || '',
    destinationOptions: [],
  };

  switch (connRes.connection?.connectionConfig?.config.case) {
    case 'pgConfig':
      return {
        ...values,
        sourceOptions: {
          haltOnNewColumnAddition: false,
        },
      };
    default:
      return values;
  }
}
