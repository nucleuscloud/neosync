'use client';
import { SingleTableEditSourceFormValues } from '@/app/(mgmt)/[account]/new/job/job-form-validations';
import {
  SchemaTable,
  getAllFormErrors,
} from '@/components/jobs/SchemaTable/SchemaTable';
import { getSchemaConstraintHandler } from '@/components/jobs/SchemaTable/schema-constraint-handler';
import { TransformerResult } from '@/components/jobs/SchemaTable/transformer-handler';
import { getTransformerFilter } from '@/components/jobs/SchemaTable/util';
import { useAccount } from '@/components/providers/account-provider';
import { Alert, AlertTitle } from '@/components/ui/alert';
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
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useGetTransformersHandler } from '@/libs/hooks/useGetTransformersHandler';
import { getErrorMessage, getTransformerFromField } from '@/util/util';
import {
  JobMappingTransformerForm,
  convertJobMappingTransformerFormToJobMappingTransformer,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import { create } from '@bufbuild/protobuf';
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  ConnectionDataService,
  ConnectionService,
  GetConnectionResponse,
  GetConnectionSchemaMapResponse,
  Job,
  JobMappingSchema,
  JobMappingTransformerSchema,
  JobService,
  ValidateJobMappingsResponse,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useQueryClient } from '@tanstack/react-query';
import { ReactElement, useEffect, useMemo, useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import {
  getSingleTableGenerateNumRows,
  toSingleTableEditGenerateJobSource,
  validateJobMapping,
} from '../../../util';
import SchemaPageSkeleton from './SchemaPageSkeleton';
import { useOnApplyDefaultClick } from './useOnApplyDefaultClick';
import { useOnImportMappings } from './useOnImportMappings';
import { useOnTransformerBulkUpdateClick } from './useOnTransformerBulkUpdateClick';
import {
  getFilteredTransformersForBulkSet,
  getOnSelectedTableToggle,
} from './util';

interface Props {
  jobId: string;
}

export default function DataGenConnectionCard({ jobId }: Props): ReactElement {
  const { account } = useAccount();

  const [validateMappingsResponse, setValidateMappingsResponse] = useState<
    ValidateJobMappingsResponse | undefined
  >();

  const [isValidatingMappings, setIsValidatingMappings] = useState(false);

  const {
    data,
    refetch: mutate,
    isLoading: isJobLoading,
  } = useQuery(JobService.method.getJob, { id: jobId }, { enabled: !!jobId });
  const { data: connectionsData, isFetching: isConnectionsValidating } =
    useQuery(
      ConnectionService.method.getConnections,
      { accountId: account?.id },
      { enabled: !!account?.id }
    );

  const connections = connectionsData?.connections ?? [];

  const form = useForm<SingleTableEditSourceFormValues>({
    resolver: yupResolver(SingleTableEditSourceFormValues),
    values: getJobSource(data?.job),
    context: { accountId: account?.id },
  });

  const fkSourceConnectionId = form.watch('source.fkSourceConnectionId');

  const {
    data: connectionSchemaDataMap,
    isLoading: isSchemaDataMapLoading,
    isFetching: isSchemaMapValidating,
  } = useQuery(
    ConnectionDataService.method.getConnectionSchemaMap,
    { connectionId: fkSourceConnectionId },
    { enabled: !!fkSourceConnectionId }
  );
  const { mutateAsync: getConnectionSchemaMapAsync } = useMutation(
    ConnectionDataService.method.getConnectionSchemaMap
  );

  const queryclient = useQueryClient();

  const { data: tableConstraints, isFetching: isTableConstraintsValidating } =
    useQuery(
      ConnectionDataService.method.getConnectionTableConstraints,
      { connectionId: fkSourceConnectionId },
      { enabled: !!fkSourceConnectionId }
    );

  const { mutateAsync: getConnectionAsync } = useMutation(
    ConnectionService.method.getConnection
  );

  const schemaConstraintHandler = useMemo(
    () =>
      getSchemaConstraintHandler(
        connectionSchemaDataMap?.schemaMap ?? {},
        tableConstraints?.primaryKeyConstraints ?? {},
        tableConstraints?.foreignKeyConstraints ?? {},
        tableConstraints?.uniqueConstraints ?? {},
        []
      ),
    [isSchemaMapValidating, isTableConstraintsValidating]
  );
  const [selectedTables, setSelectedTables] = useState<Set<string>>(new Set());
  const { append, remove, fields, update } =
    useFieldArray<SingleTableEditSourceFormValues>({
      control: form.control,
      name: 'mappings',
    });

  useEffect(() => {
    if (isJobLoading || isSchemaDataMapLoading || selectedTables.size > 0) {
      return;
    }
    const js = getJobSource(data?.job);
    setSelectedTables(
      new Set(
        js.mappings.map((mapping) => `${mapping.schema}.${mapping.table}`)
      )
    );

    // handle missing columns
  }, [isJobLoading, isSchemaDataMapLoading]);

  useEffect(() => {
    const connSchemaMap = connectionSchemaDataMap?.schemaMap;
    if (isJobLoading || isSchemaMapValidating || !connSchemaMap) {
      return;
    }
    const existingCols: Record<string, Set<string>> = {};
    fields.forEach((mapping) => {
      const key = `${mapping.schema}.${mapping.table}`;
      const uniqcols = existingCols[key];
      if (uniqcols) {
        uniqcols.add(mapping.column);
      } else {
        existingCols[key] = new Set([mapping.column]);
      }
    });
    const toAdd: SingleTableEditSourceFormValues['mappings'] = [];
    Object.entries(existingCols).forEach(([key, currcols]) => {
      const schemaResp = connSchemaMap[key];
      if (!schemaResp) {
        return;
      }
      schemaResp.schemas.forEach((dbcol) => {
        if (!currcols.has(dbcol.column)) {
          toAdd.push({
            schema: dbcol.schema,
            table: dbcol.table,
            column: dbcol.column,
            transformer: convertJobMappingTransformerToForm(
              create(JobMappingTransformerSchema)
            ),
          });
        }
      });
    });
    if (toAdd.length > 0) {
      // must be set instead of append as sometimes this is triggered twice and would result in duplicate values being inserted.
      form.setValue('mappings', [...fields, ...toAdd]);
    }
  }, [isJobLoading, isSchemaMapValidating]);

  const connectionsMap = useMemo(
    () => new Map(connections.map((c) => [c.id, c])),
    [isConnectionsValidating]
  );

  const { mutateAsync: updateJobSrcConnection } = useMutation(
    JobService.method.updateJobSourceConnection
  );

  useEffect(() => {
    if (!fkSourceConnectionId || !account?.id) {
      return;
    }
    const validateJobMappings = async () => {
      await validateMappings();
    };
    validateJobMappings();
  }, [selectedTables, fkSourceConnectionId, account?.id]);

  const { mutateAsync: validateJobMappingsAsync } = useMutation(
    JobService.method.validateJobMappings
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
    setTransformer(idx, transformer) {
      form.setValue(`mappings.${idx}.transformer`, transformer, {
        shouldDirty: true,
        shouldTouch: true,
        shouldValidate: false,
      });
    },
    triggerUpdate() {
      form.trigger('mappings');
    },
    setSelectedTables: setSelectedTables,
  });

  function getAvailableTransformers(idx: number): TransformerResult {
    const row = fields[idx];
    return handler.getFilteredTransformers(
      getTransformerFilter(
        schemaConstraintHandler,
        {
          schema: row.schema,
          table: row.table,
          column: row.column,
        },
        'generate'
      )
    );
  }

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

  if (isJobLoading || isSchemaDataMapLoading || isGetTransformersLoading) {
    return <SchemaPageSkeleton />;
  }

  async function onSubmit(values: SingleTableEditSourceFormValues) {
    const job = data?.job;
    if (!job) {
      return;
    }
    try {
      await updateJobSrcConnection({
        id: job.id,
        mappings: values.mappings.map((m) => {
          return create(JobMappingSchema, {
            schema: m.schema,
            table: m.table,
            column: m.column,
            transformer:
              convertJobMappingTransformerFormToJobMappingTransformer(
                m.transformer
              ),
          });
        }),
        source: toSingleTableEditGenerateJobSource(values),
      });
      toast.success('Successfully updated job source connection!');
      mutate();
    } catch (err) {
      console.error(err);
      toast.error('Unable to update job source connection', {
        description: getErrorMessage(err),
      });
    }
  }

  async function validateMappings() {
    if (!fkSourceConnectionId || form.getValues('mappings').length == 0) {
      return;
    }
    try {
      setIsValidatingMappings(true);
      const res = await validateJobMapping(
        fields,
        account?.id || '',
        [],
        validateJobMappingsAsync,
        toSingleTableEditGenerateJobSource(form.getValues())
      );
      setValidateMappingsResponse(res);
      form.trigger('mappings');
    } catch (error) {
      console.error('Failed to validate job mappings:', error);
      toast.error('Unable to validate job mappings', {
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
    fields,
    remove,
    append
  );

  async function onTableConstraintSourceChange(value: string): Promise<void> {
    try {
      const newValues = await getUpdatedValues(
        value,
        form.getValues(),
        async (id) => {
          const resp = await getConnectionAsync({ id });
          queryclient.setQueryData(
            createConnectQueryKey({
              schema: ConnectionService.method.getConnection,
              input: { id },
              cardinality: undefined,
            }),
            resp
          );
          return resp;
        },
        async (id) => {
          const resp = await getConnectionSchemaMapAsync({ connectionId: id });
          queryclient.setQueryData(
            createConnectQueryKey({
              schema: ConnectionDataService.method.getConnectionSchemaMap,
              input: { connectionId: id },
              cardinality: undefined,
            }),
            resp
          );
          return resp;
        }
      );
      form.reset(newValues);
      const newMapping =
        newValues.mappings.length > 0 ? newValues.mappings[0] : undefined;
      if (newMapping && newMapping.schema && newMapping.table) {
        setSelectedTables(
          new Set([`${newMapping.schema}.${newMapping.table}`])
        );
      } else {
        setSelectedTables(new Set());
      }
    } catch (err) {
      form.reset({
        ...form.getValues(),
        source: { ...form.getValues('source'), fkSourceConnectionId: value },
      });
      toast.error(
        'Unable to get connection schema on table constraint source change.',
        {
          description: getErrorMessage(err),
        }
      );
    }
  }

  const fkConn = connections.find((c) => c.id === fkSourceConnectionId);

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <FormField
          control={form.control}
          name="source.fkSourceConnectionId"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Table Schema Connection</FormLabel>
              <FormDescription>
                Connection used for table schema. Must be of the same type as
                the destination.
              </FormDescription>
              <FormControl>
                <Select
                  value={field.value}
                  onValueChange={async (value) => {
                    if (!value) {
                      return;
                    }
                    field.onChange(value);
                    await onTableConstraintSourceChange(value);
                  }}
                >
                  <SelectTrigger>
                    <SelectValue placeholder={fkConn?.name} />
                  </SelectTrigger>
                  <SelectContent>
                    {connections
                      .filter((c) => {
                        const dests = data?.job?.destinations ?? [];

                        return (
                          c.connectionConfig?.config.case !== 'awsS3Config' &&
                          c.connectionConfig?.config.case !==
                            'gcpCloudstorageConfig' &&
                          dests.some((dest) => {
                            const destConn = connectionsMap.get(
                              dest.connectionId
                            );
                            return (
                              !!destConn &&
                              destConn.connectionConfig?.config.case ===
                                c.connectionConfig?.config.case
                            );
                          })
                        );
                      })
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

        <FormField
          control={form.control}
          name="numRows"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Number of Rows</FormLabel>
              <FormDescription>The number of rows to generate.</FormDescription>
              <FormControl>
                <Input
                  {...field}
                  type="number"
                  onChange={(e) => {
                    const numberValue = e.target.valueAsNumber;
                    if (!isNaN(numberValue)) {
                      field.onChange(numberValue);
                    }
                  }}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <SchemaTable
          data={fields}
          jobType="generate"
          constraintHandler={schemaConstraintHandler}
          schema={connectionSchemaDataMap?.schemaMap ?? {}}
          isSchemaDataReloading={isSchemaMapValidating}
          selectedTables={selectedTables}
          onSelectedTableToggle={onSelectedTableToggle}
          formErrors={getAllFormErrors(
            form.formState.errors,
            fields,
            validateMappingsResponse
          )}
          isJobMappingsValidating={isValidatingMappings}
          onValidate={validateMappings}
          onImportMappingsClick={onImportMappingsClick}
          onTransformerUpdate={(idx, cfg) => {
            onTransformerUpdate(idx, cfg);
          }}
          getAvailableTransformers={getAvailableTransformers}
          getTransformerFromField={(idx) => {
            const row = fields[idx];
            return getTransformerFromField(handler, row.transformer);
          }}
          getAvailableTransformersForBulk={(rows) => {
            return getFilteredTransformersForBulkSet(
              rows,
              handler,
              schemaConstraintHandler,
              'generate',
              'relational'
            );
          }}
          getTransformerFromFieldValue={(fvalue) => {
            return getTransformerFromField(handler, fvalue);
          }}
          onApplyDefaultClick={onApplyDefaultClick}
          onTransformerBulkUpdate={onTransformerBulkUpdate}
        />

        {form.formState.errors.mappings && (
          <Alert variant="destructive">
            <AlertTitle className="flex flex-row space-x-2 justify-center">
              <ExclamationTriangleIcon />
              <p>Please fix form errors and try again.</p>
            </AlertTitle>
          </Alert>
        )}
        <div className="flex flex-row gap-1 justify-end">
          <Button key="submit" type="submit">
            Update
          </Button>
        </div>
      </form>
    </Form>
  );
}

function getJobSource(job?: Job): SingleTableEditSourceFormValues {
  if (!job) {
    return {
      source: {
        fkSourceConnectionId: '',
      },
      mappings: [],
      numRows: 0,
    };
  }

  const numRows =
    job.source?.options?.config.case === 'generate'
      ? getSingleTableGenerateNumRows(job.source?.options?.config.value)
      : 0;

  const mappings: SingleTableEditSourceFormValues['mappings'] = (
    job.mappings ?? []
  ).map((mapping) => {
    return {
      schema: mapping.schema,
      table: mapping.table,
      column: mapping.column,
      transformer: mapping.transformer
        ? convertJobMappingTransformerToForm(mapping.transformer)
        : convertJobMappingTransformerToForm(
            create(JobMappingTransformerSchema)
          ),
    };
  });

  let fkSourceConnectionId = '';
  if (job.source?.options?.config.case === 'generate') {
    fkSourceConnectionId =
      job.source.options.config.value.fkSourceConnectionId ?? '';
  }

  return {
    source: {
      fkSourceConnectionId: fkSourceConnectionId,
    },
    mappings: mappings,
    numRows: numRows,
  };
}

async function getUpdatedValues(
  connectionId: string,
  originalValues: SingleTableEditSourceFormValues,
  getConnectionById: (id: string) => Promise<GetConnectionResponse>,
  getConnectionSchemaMapAsync: (
    id: string
  ) => Promise<GetConnectionSchemaMapResponse>
): Promise<SingleTableEditSourceFormValues> {
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

  return {
    ...originalValues,
    source: {
      fkSourceConnectionId: connectionId,
    },
    mappings,
  };
}
