'use client';
import SourceOptionsForm from '@/components/jobs/Form/SourceOptionsForm';
import NosqlTable from '@/components/jobs/NosqlTable/NosqlTable';
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
import {
  ConnectionSchemaMap,
  GetConnectionSchemaMapResponse,
  getConnectionSchema,
  useGetConnectionSchemaMap,
} from '@/libs/hooks/useGetConnectionSchemaMap';
import { validateJobMapping } from '@/libs/requests/validateJobMappings';
import { getErrorMessage } from '@/util/util';
import {
  SchemaFormValues,
  SourceFormValues,
  VirtualForeignConstraintFormValues,
  convertJobMappingTransformerFormToJobMappingTransformer,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  Connection,
  GetConnectionResponse,
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
  getConnectionTableConstraints,
  getConnections,
  getJob,
  updateJobSourceConnection,
} from '@neosync/sdk/connectquery';
import { ReactElement, useEffect, useMemo, useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import * as Yup from 'yup';
import SchemaPageSkeleton from './SchemaPageSkeleton';
import { getOnSelectedTableToggle } from './util';

interface Props {
  jobId: string;
}

const FORM_SCHEMA = SourceFormValues.concat(
  Yup.object({
    destinationIds: Yup.array().of(Yup.string().required()),
  })
).concat(SchemaFormValues);
type SourceFormValues = Yup.InferType<typeof FORM_SCHEMA>;

function getConnectionIdFromSource(
  js: JobSource | undefined
): string | undefined {
  if (
    js?.options?.config.case === 'postgres' ||
    js?.options?.config.case === 'mysql' ||
    js?.options?.config.case === 'awsS3' ||
    js?.options?.config.case === 'mongodb'
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
    isValidating: isSchemaMapValidating,
    mutate: mutateGetConnectionSchemaMap,
  } = useGetConnectionSchemaMap(account?.id ?? '', sourceConnectionId ?? '');

  const { isLoading: isConnectionsLoading, data: connectionsData } = useQuery(
    getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const connections = connectionsData?.connections ?? [];

  const { mutateAsync: updateJobSrcConnection } = useMutation(
    updateJobSourceConnection
  );

  const [validateMappingsResponse, setValidateMappingsResponse] = useState<
    ValidateJobMappingsResponse | undefined
  >();

  const [isValidatingMappings, setIsValidatingMappings] = useState(false);

  const form = useForm<SourceFormValues>({
    resolver: yupResolver<SourceFormValues>(FORM_SCHEMA),
    values: getJobSource(data?.job, connectionSchemaDataMap?.schemaMap),
    context: { accountId: account?.id },
  });
  const formVirtualForeignKeys = form.watch('virtualForeignKeys');

  const { data: tableConstraints, isFetching: isTableConstraintsValidating } =
    useQuery(
      getConnectionTableConstraints,
      { connectionId: sourceConnectionId },
      { enabled: !!sourceConnectionId }
    );
  const { mutateAsync: getConnectionAsync } = useMutation(getConnection);

  const schemaConstraintHandler = useMemo(() => {
    const virtualForeignKeys = data?.job?.virtualForeignKeys ?? [];
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
    formVirtualForeignKeys,
  ]);
  const [selectedTables, setSelectedTables] = useState<Set<string>>(new Set());

  const { append, remove, update, fields } = useFieldArray<SourceFormValues>({
    control: form.control,
    name: 'mappings',
  });

  const { append: appendVfk, remove: removeVfk } =
    useFieldArray<SourceFormValues>({
      control: form.control,
      name: 'virtualForeignKeys',
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

  async function onSourceChange(value: string): Promise<void> {
    try {
      const newValues = await getUpdatedValues(
        account?.id ?? '',
        value,
        form.getValues(),
        (id) => getConnectionAsync({ id }),
        mutateGetConnectionSchemaMap
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

  async function onSubmit(values: SourceFormValues) {
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
      mutate();
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to update job source connection',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }
  const formMappings = form.watch('mappings');
  async function validateMappings() {
    try {
      setIsValidatingMappings(true);
      const res = await validateJobMapping(
        sourceConnectionId || '',
        formMappings,
        account?.id || '',
        formVirtualForeignKeys
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
        vfks
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
    fields,
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

  if (isConnectionsLoading || isSchemaDataMapLoading || isJobDataLoading) {
    return <SchemaPageSkeleton />;
  }

  const source = connections.find((item) => item.id === sourceConnectionId);
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
                            !form.getValues().destinationIds?.includes(c.id) &&
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
              }}
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
  values: SourceFormValues,
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

function getJobSource(
  job?: Job,
  connSchemaMap?: ConnectionSchemaMap
): SourceFormValues {
  if (!job || !connSchemaMap) {
    return {
      sourceId: '',
      sourceOptions: {
        haltOnNewColumnAddition: false,
      },
      destinationIds: [],
      mappings: [],
      virtualForeignKeys: [],
      connectionId: '',
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
      dbcols.forEach((dbcol) => {
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

  const destinationIds = job?.destinations.map((d) => d.connectionId);
  const values = {
    sourceOptions: {},
    destinationIds: destinationIds,
    mappings: mappings || [],
    virtualForeignKeys: virtualForeignKeys || [],
  };
  const yupValidationValues = {
    ...values,
    sourceId: getConnectionIdFromSource(job.source) || '',
    mappings,
    connectionId: getConnectionIdFromSource(job.source) || '',
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
    default:
      return yupValidationValues;
  }
}

async function getUpdatedValues(
  accountId: string,
  connectionId: string,
  originalValues: SourceFormValues,
  getConnectionById: (id: string) => Promise<GetConnectionResponse>,
  mutateConnectionSchemaResponse: (
    schemaRes: GetConnectionSchemaMapResponse
  ) => void
): Promise<SourceFormValues> {
  const [schemaRes, connRes] = await Promise.all([
    getConnectionSchema(accountId, connectionId),
    getConnectionById(connectionId),
  ]);

  if (!schemaRes || !connRes) {
    return originalValues;
  }

  const sameKeys = new Set(
    Object.values(schemaRes.schemaMap).flatMap((dbcols) =>
      dbcols.map((dbcol) => `${dbcol.schema}.${dbcol.table}.${dbcol.column}`)
    )
  );

  const mappings = originalValues.mappings.filter((mapping) =>
    sameKeys.has(`${mapping.schema}.${mapping.table}.${mapping.column}`)
  );

  const values = {
    sourceId: connectionId || '',
    sourceOptions: {},
    destinationIds: originalValues.destinationIds,
    mappings,
    connectionId: connectionId || '',
  };
  mutateConnectionSchemaResponse(schemaRes);

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

function isNosqlSource(connection: Connection): boolean {
  switch (connection.connectionConfig?.config.case) {
    case 'mongoConfig':
      return true;
    default: {
      return false;
    }
  }
}
