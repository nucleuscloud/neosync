'use client';
import ConnectionSelectContent from '@/app/(mgmt)/[account]/new/job/connect/ConnectionSelectContent';
import SourceOptionsForm from '@/components/jobs/Form/SourceOptionsForm';
import NosqlTable from '@/components/jobs/NosqlTable/NosqlTable';
import { OnTableMappingUpdateRequest } from '@/components/jobs/NosqlTable/TableMappings/Columns';
import {
  SchemaTable,
  getAllFormErrors,
} from '@/components/jobs/SchemaTable/SchemaTable';
import { getSchemaConstraintHandler } from '@/components/jobs/SchemaTable/schema-constraint-handler';
import { TransformerResult } from '@/components/jobs/SchemaTable/transformer-handler';
import {
  getTransformerFilter,
  splitCollection,
} from '@/components/jobs/SchemaTable/util';
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
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useGetTransformersHandler } from '@/libs/hooks/useGetTransformersHandler';
import { splitConnections } from '@/libs/utils';
import { getErrorMessage, getTransformerFromField } from '@/util/util';
import {
  DataSyncSourceFormValues,
  EditDestinationOptionsFormValues,
  JobMappingFormValues,
  JobMappingTransformerForm,
  VirtualForeignConstraintFormValues,
  convertJobMappingTransformerFormToJobMappingTransformer,
  convertJobMappingTransformerToForm,
  toJobSourcePostgresNewColumnAdditionStrategy,
  toNewColumnAdditionStrategy,
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
  DynamoDBSourceUnmappedTransformConfig,
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
  MssqlSourceConnectionOptions,
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
import { toast } from 'sonner';
import {
  getDefaultUnmappedTransformConfig,
  toDynamoDbSourceUnmappedOptionsFormValues,
  validateJobMapping,
} from '../../../util';
import SchemaPageSkeleton from './SchemaPageSkeleton';
import { useOnApplyDefaultClick } from './useOnApplyDefaultClick';
import { useOnImportMappings } from './useOnImportMappings';
import { useOnTransformerBulkUpdateClick } from './useOnTransformerBulkUpdateClick';
import {
  getConnectionIdFromSource,
  getDestinationDetailsRecord,
  getDynamoDbDestinations,
  getFilteredTransformersForBulkSet,
  getOnSelectedTableToggle,
  isDynamoDBConnection,
  isNosqlSource,
  shouldShowDestinationTableMappings,
} from './util';

interface Props {
  jobId: string;
}

export default function DataSyncConnectionCard({ jobId }: Props): ReactElement {
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
      toast.error('Unable to get connection schema', {
        description: getErrorMessage(err),
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
      toast.success('Successfully updated source connection!');
      // hold off on mutating until after we update the job dest connections for dynamo conns
      if (connection.connectionConfig?.config.case !== 'dynamodbConfig') {
        mutate();
      }
    } catch (err) {
      console.error(err);
      toast.error('Unable to update job source connnection', {
        description: getErrorMessage(err),
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
          if (!destOpts.dynamodb) {
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
                  tableMappings: destOpts.dynamodb.tableMappings ?? [],
                },
              },
            },
          });
        })
      );
      toast.success('Successfully updated job destination connection(s)');
      mutate();
    } catch (err) {
      console.error(err);
      toast.error('Unable to update one or all job destination connections', {
        description: getErrorMessage(err),
      });
    }
  }

  async function validateMappings(
    mappings: JobMappingFormValues[] = formMappings
  ) {
    try {
      setIsValidatingMappings(true);
      const res = await validateJobMapping(
        sourceConnectionId || '',
        mappings,
        account?.id || '',
        formVirtualForeignKeys,
        validateJobMappingsAsync
      );
      setValidateMappingsResponse(res);
    } catch (error) {
      console.error('Failed to validate job mappings:', error);
      toast.error('Unable to validate job mappings', {
        description: getErrorMessage(error),
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
      toast.error('Unable to validate virtual foreign keys', {
        description: getErrorMessage(error),
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
      const tm = destOpt?.dynamodb?.tableMappings.find(
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

  const { handler, isLoading: isGetTransformersLoading } =
    useGetTransformersHandler(account?.id ?? '');

  function onTransformerUpdate(
    index: number,
    transformer: JobMappingTransformerForm
  ): void {
    const val = form.getValues(`mappings.${index}`);
    update(index, {
      schema: val.schema,
      table: val.table,
      column: val.column,
      transformer,
    });
  }

  const { onClick: onImportMappingsClick } = useOnImportMappings({
    setMappings(mappings) {
      form.setValue('mappings', mappings, {
        shouldDirty: true,
        shouldTouch: true,
        shouldValidate: false,
      });
    },
    getMappings() {
      return form.getValues('mappings');
    },
    appendNewMappings(mappings) {
      append(mappings);
    },
    setTransformer: onTransformerUpdate,
    async triggerUpdate() {
      await form.trigger('mappings');
      setTimeout(() => {
        // using form.getvalues instead of formMappings as it is more up to date for some reason (bug?)
        validateMappings(form.getValues('mappings'));
      }, 0);
    },
    setSelectedTables: setSelectedTables,
  });

  const { onClick: onApplyDefaultClick } = useOnApplyDefaultClick({
    getMappings() {
      return form.getValues('mappings');
    },
    setMappings(mappings) {
      form.setValue('mappings', mappings, {
        shouldDirty: true,
        shouldTouch: true,
        shouldValidate: false,
      });
    },
    constraintHandler: schemaConstraintHandler,
    triggerUpdate() {
      form.trigger('mappings');
    },
  });

  const { onClick: onTransformerBulkUpdate } = useOnTransformerBulkUpdateClick({
    getMappings() {
      return form.getValues('mappings');
    },
    setMappings(mappings) {
      form.setValue('mappings', mappings, {
        shouldDirty: true,
        shouldTouch: true,
        shouldValidate: false,
      });
    },
    triggerUpdate() {
      form.trigger('mappings');
    },
  });

  if (
    isConnectionsLoading ||
    isSchemaDataMapLoading ||
    isJobDataLoading ||
    isGetTransformersLoading
  ) {
    return <SchemaPageSkeleton />;
  }

  const source = connectionsRecord[sourceConnectionId ?? ''] as
    | Connection
    | undefined;

  const dynamoDBDestinations = getDynamoDbDestinations(
    data?.job?.destinations ?? []
  );

  const splitSourceConnections = splitConnections(
    connections.filter(
      (c) =>
        !data?.job?.destinations.map((d) => d.connectionId)?.includes(c.id) &&
        c.connectionConfig?.config.case !== 'awsS3Config' &&
        c.connectionConfig?.config.case !== 'openaiConfig' &&
        c.connectionConfig?.config.case !== 'gcpCloudstorageConfig'
    )
  );

  function getAvailableTransformers(idx: number): TransformerResult {
    const row = formMappings[idx];
    return handler.getFilteredTransformers(
      getTransformerFilter(
        schemaConstraintHandler,
        {
          schema: row.schema,
          table: row.table,
          column: row.column,
        },
        'sync'
      )
    );
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        <div className="space-y-4">
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
                      <ConnectionSelectContent
                        postgres={splitSourceConnections.postgres}
                        mysql={splitSourceConnections.mysql}
                        mongodb={splitSourceConnections.mongodb}
                        dynamodb={splitSourceConnections.dynamodb}
                        mssql={splitSourceConnections.mssql}
                      />
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
            value={form.watch('sourceOptions')}
            setValue={(newOpts) => {
              form.setValue('sourceOptions', newOpts, {
                shouldDirty: true,
                shouldTouch: true,
                shouldValidate: true,
              });
            }}
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
              onRemoveMappings={(indices) => {
                const indexSet = new Set(indices);
                const remainingTables = formMappings
                  .filter((_, idx) => !indexSet.has(idx))
                  .map((fm) => fm.table);

                if (indices.length > 0) {
                  remove(indices);
                }

                if (!source || isDynamoDBConnection(source)) {
                  return;
                }

                // Check and update destinationOptions if needed
                const destOpts = form.getValues('destinationOptions');
                const updatedDestOpts = destOpts
                  .map((opt) => {
                    if (opt.dynamodb) {
                      const updatedTableMappings =
                        opt.dynamodb.tableMappings.filter((tm) => {
                          // Check if any columns remain for the table
                          const tableColumnsExist = remainingTables.some(
                            (table) => table === tm.sourceTable
                          );
                          return tableColumnsExist;
                        });

                      return {
                        ...opt,
                        dynamoDb: {
                          ...opt.dynamodb,
                          tableMappings: updatedTableMappings,
                        },
                      };
                    }
                    return opt;
                  })
                  .filter(
                    (opt) => (opt.dynamodb?.tableMappings.length ?? 0) > 0
                  );
                form.setValue('destinationOptions', updatedDestOpts);
              }}
              onEditMappings={(values, index) => {
                if (index >= 0 && index < formMappings.length) {
                  update(index, values);
                }
              }}
              onAddMappings={(values) => {
                append(
                  values.map((v) => {
                    const [schema, table] = splitCollection(v.collection);
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
                  (dest): EditDestinationOptionsFormValues => {
                    const opt = existing.get(dest.id);
                    if (opt) {
                      const sourceSet = new Set(
                        opt.dynamodb?.tableMappings.map(
                          (mapping) => mapping.sourceTable
                        ) ?? []
                      );

                      // Add missing uniqueCollections to the existing tableMappings
                      const updatedTableMappings = [
                        ...(opt.dynamodb?.tableMappings ?? []),
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
                        dynamodb: {
                          ...opt.dynamodb,
                          tableMappings: updatedTableMappings,
                        },
                      };
                    }

                    return {
                      destinationId: dest.id,
                      dynamodb: {
                        tableMappings: uniqueCollections.map((c) => {
                          const [, table] = c.split('.');
                          return {
                            sourceTable: table,
                            destinationTable: '',
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
              onImportMappingsClick={onImportMappingsClick}
              getAvailableTransformers={getAvailableTransformers}
              getTransformerFromField={(idx) => {
                const row = formMappings[idx];
                return getTransformerFromField(handler, row.transformer);
              }}
              getAvailableTransformersForBulk={(rows) => {
                return getFilteredTransformersForBulkSet(
                  rows,
                  handler,
                  schemaConstraintHandler,
                  'sync',
                  'nosql'
                );
              }}
              getTransformerFromFieldValue={(fvalue) => {
                return getTransformerFromField(handler, fvalue);
              }}
              onApplyDefaultClick={onApplyDefaultClick}
              onTransformerBulkUpdate={onTransformerBulkUpdate}
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
              onImportMappingsClick={onImportMappingsClick}
              onTransformerUpdate={(idx, cfg) => {
                onTransformerUpdate(idx, cfg);
              }}
              getAvailableTransformers={getAvailableTransformers}
              getTransformerFromField={(idx) => {
                const row = formMappings[idx];
                return getTransformerFromField(handler, row.transformer);
              }}
              getAvailableTransformersForBulk={(rows) => {
                return getFilteredTransformersForBulkSet(
                  rows,
                  handler,
                  schemaConstraintHandler,
                  'sync',
                  'relational'
                );
              }}
              getTransformerFromFieldValue={(fvalue) => {
                return getTransformerFromField(handler, fvalue);
              }}
              onApplyDefaultClick={onApplyDefaultClick}
              onTransformerBulkUpdate={onTransformerBulkUpdate}
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
    case 'pgConfig': {
      return new JobSourceOptions({
        config: {
          case: 'postgres',
          value: new PostgresSourceConnectionOptions({
            ...getExistingPostgresSourceConnectionOptions(job),
            connectionId: newSourceId,
            newColumnAdditionStrategy:
              toJobSourcePostgresNewColumnAdditionStrategy(
                values.sourceOptions.postgres?.newColumnAdditionStrategy
              ),
          }),
        },
      });
    }
    case 'mysqlConfig':
      return new JobSourceOptions({
        config: {
          case: 'mysql',
          value: new MysqlSourceConnectionOptions({
            ...getExistingMysqlSourceConnectionOptions(job),
            connectionId: newSourceId,
            haltOnNewColumnAddition:
              values.sourceOptions.mysql?.haltOnNewColumnAddition,
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
            unmappedTransforms: new DynamoDBSourceUnmappedTransformConfig({
              b: values.sourceOptions.dynamodb?.unmappedTransformConfig?.byte
                ? convertJobMappingTransformerFormToJobMappingTransformer(
                    values.sourceOptions.dynamodb.unmappedTransformConfig.byte
                  )
                : undefined,
              boolean: values.sourceOptions.dynamodb?.unmappedTransformConfig
                ?.boolean
                ? convertJobMappingTransformerFormToJobMappingTransformer(
                    values.sourceOptions.dynamodb.unmappedTransformConfig
                      .boolean
                  )
                : undefined,
              n: values.sourceOptions.dynamodb?.unmappedTransformConfig?.n
                ? convertJobMappingTransformerFormToJobMappingTransformer(
                    values.sourceOptions.dynamodb.unmappedTransformConfig.n
                  )
                : undefined,
              s: values.sourceOptions.dynamodb?.unmappedTransformConfig?.s
                ? convertJobMappingTransformerFormToJobMappingTransformer(
                    values.sourceOptions.dynamodb.unmappedTransformConfig.s
                  )
                : undefined,
            }),
            enableConsistentRead:
              values.sourceOptions.dynamodb?.enableConsistentRead,
          }),
        },
      });
    }
    case 'mssqlConfig': {
      return new JobSourceOptions({
        config: {
          case: 'mssql',
          value: new MssqlSourceConnectionOptions({
            ...getExistingMssqlSourceConnectionOptions(job),
            connectionId: newSourceId,
            haltOnNewColumnAddition:
              values.sourceOptions.mssql?.haltOnNewColumnAddition,
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

function getExistingMssqlSourceConnectionOptions(
  job: Job
): MssqlSourceConnectionOptions | undefined {
  return job.source?.options?.config.case === 'mssql'
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
      sourceOptions: {},
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
    job.source?.options?.config.case === 'mysql' ||
    job.source?.options?.config.case === 'mssql'
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
          postgres: {
            newColumnAdditionStrategy: toNewColumnAdditionStrategy(
              job.source.options.config.value.newColumnAdditionStrategy
            ),
          },
        },
      };
    case 'mysql':
      return {
        ...yupValidationValues,
        sourceId: getConnectionIdFromSource(job.source) || '',
        sourceOptions: {
          mysql: {
            haltOnNewColumnAddition:
              job?.source?.options?.config.value.haltOnNewColumnAddition,
          },
        },
      };
    case 'mongodb':
      return {
        ...yupValidationValues,
        sourceId: getConnectionIdFromSource(job.source) || '',
        sourceOptions: {},
      };
    case 'dynamodb': {
      const destOpts: EditDestinationOptionsFormValues[] = [];
      job.destinations.forEach((d) => {
        if (d.options?.config.case !== 'dynamodbOptions') {
          return;
        }
        destOpts.push({
          destinationId: d.id,
          dynamodb: {
            tableMappings: d.options.config.value.tableMappings ?? [],
          },
        });
      });
      return {
        ...yupValidationValues,
        sourceId: getConnectionIdFromSource(job.source) || '',
        sourceOptions: {
          dynamodb: {
            unmappedTransformConfig: toDynamoDbSourceUnmappedOptionsFormValues(
              job.source?.options?.config?.value.unmappedTransforms
            ),
            enableConsistentRead:
              job.source.options.config.value.enableConsistentRead,
          },
        },
        destinationOptions: destOpts,
      };
    }
    case 'mssql': {
      return {
        ...yupValidationValues,
        sourceId: getConnectionIdFromSource(job.source) || '',
        sourceOptions: {
          mssql: {
            haltOnNewColumnAddition:
              job?.source?.options?.config.value.haltOnNewColumnAddition,
          },
        },
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
    mappings,
    connectionId: connectionId || '',
    destinationOptions: [],
  };

  switch (connRes.connection?.connectionConfig?.config.case) {
    case 'pgConfig':
      return {
        ...values,
        sourceOptions: {
          postgres: {
            newColumnAdditionStrategy: 'halt',
          },
        },
      };
    case 'mysqlConfig': {
      return {
        ...values,
        sourceOptions: {
          mysql: {
            haltOnNewColumnAddition: false,
          },
        },
      };
    }
    case 'dynamodbConfig': {
      return {
        ...values,
        sourceOptions: {
          dynamodb: {
            unmappedTransformConfig: getDefaultUnmappedTransformConfig(),
            enableConsistentRead: false,
          },
        },
      };
    }
    case 'mssqlConfig': {
      return {
        ...values,
        sourceOptions: {
          mssql: {
            haltOnNewColumnAddition: false,
          },
        },
      };
    }
    default:
      return values;
  }
}
