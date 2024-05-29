'use client';
import { SingleTableEditSourceFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import {
  SchemaTable,
  getAllFormErrors,
} from '@/components/jobs/SchemaTable/SchemaTable';
import { getSchemaConstraintHandler } from '@/components/jobs/SchemaTable/schema-constraint-handler';
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
import { useToast } from '@/components/ui/use-toast';
import { getConnection } from '@/libs/hooks/useGetConnection';
import {
  GetConnectionSchemaMapResponse,
  getConnectionSchema,
  useGetConnectionSchemaMap,
} from '@/libs/hooks/useGetConnectionSchemaMap';
import { useGetConnectionTableConstraints } from '@/libs/hooks/useGetConnectionTableConstraints';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { validateJobMapping } from '@/libs/requests/validateJobMappings';
import { getErrorMessage } from '@/util/util';
import {
  convertJobMappingTransformerFormToJobMappingTransformer,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  GenerateSourceOptions,
  GenerateSourceSchemaOption,
  GenerateSourceTableOption,
  Job,
  JobMapping,
  JobMappingTransformer,
  JobSource,
  JobSourceOptions,
  UpdateJobSourceConnectionRequest,
  UpdateJobSourceConnectionResponse,
  ValidateJobMappingsResponse,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ReactElement, useEffect, useMemo, useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import { KeyedMutator } from 'swr';
import SchemaPageSkeleton from './SchemaPageSkeleton';
import { getOnSelectedTableToggle } from './util';

interface Props {
  jobId: string;
}

export default function DataGenConnectionCard({ jobId }: Props): ReactElement {
  const { toast } = useToast();
  const { account } = useAccount();

  const [validateMappingsResponse, setValidateMappingsResponse] = useState<
    ValidateJobMappingsResponse | undefined
  >();

  const [isValidatingMappings, setIsValidatingMappings] = useState(false);

  const {
    data,
    mutate,
    isLoading: isJobLoading,
  } = useGetJob(account?.id ?? '', jobId);

  const { data: connectionsData, isValidating: isConnectionsValidating } =
    useGetConnections(account?.id ?? '');

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
    isValidating: isSchemaMapValidating,
    mutate: mutateGetConnectionSchemaMap,
  } = useGetConnectionSchemaMap(account?.id ?? '', fkSourceConnectionId ?? '');

  const { data: tableConstraints, isValidating: isTableConstraintsValidating } =
    useGetConnectionTableConstraints(
      account?.id ?? '',
      fkSourceConnectionId ?? ''
    );

  const schemaConstraintHandler = useMemo(
    () =>
      getSchemaConstraintHandler(
        connectionSchemaDataMap?.schemaMap ?? {},
        tableConstraints?.primaryKeyConstraints ?? {},
        tableConstraints?.foreignKeyConstraints ?? {},
        tableConstraints?.uniqueConstraints ?? {}
      ),
    [isSchemaMapValidating, isTableConstraintsValidating]
  );
  const [selectedTables, setSelectedTables] = useState<Set<string>>(new Set());
  const { append, remove, fields } =
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
      const dbcols = connSchemaMap[key];
      if (!dbcols) {
        return;
      }
      dbcols.forEach((dbcol) => {
        if (!currcols.has(dbcol.column)) {
          toAdd.push({
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
    if (toAdd.length > 0) {
      // must be set instead of append as sometimes this is triggered twice and would result in duplicate values being inserted.
      form.setValue('mappings', [...fields, ...toAdd]);
    }
  }, [isJobLoading, isSchemaMapValidating]);

  const connectionsMap = useMemo(
    () => new Map(connections.map((c) => [c.id, c])),
    [isConnectionsValidating]
  );

  useEffect(() => {
    const validateJobMappings = async () => {
      await validateMappings();
    };
    validateJobMappings();
  }, [selectedTables]);

  if (isJobLoading || isSchemaDataMapLoading) {
    return <SchemaPageSkeleton />;
  }

  async function onSubmit(values: SingleTableEditSourceFormValues) {
    const job = data?.job;
    if (!job) {
      return;
    }
    try {
      await updateJobConnection(account?.id ?? '', job, values);
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

  async function validateMappings() {
    try {
      setIsValidatingMappings(true);
      const res = await validateJobMapping(
        fkSourceConnectionId || '',
        fields,
        account?.id || ''
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
        account?.id ?? '',
        value,
        form.getValues(),
        mutateGetConnectionSchemaMap
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
      toast({
        title:
          'Unable to get connection schema on table constraint source change.',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
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
  let numRows = 0;
  if (job.source?.options?.config.case === 'generate') {
    const srcSchemas = job.source.options.config.value.schemas;
    if (srcSchemas.length > 0) {
      const tables = srcSchemas[0].tables;
      if (tables.length > 0) {
        numRows = Number(tables[0].rowCount); // this will be an issue if the number is bigger than what js allows
      }
    }
  }

  const mappings: SingleTableEditSourceFormValues['mappings'] = (
    job.mappings ?? []
  ).map((mapping) => {
    return {
      schema: mapping.schema,
      table: mapping.table,
      column: mapping.column,
      transformer: mapping.transformer
        ? convertJobMappingTransformerToForm(mapping.transformer)
        : convertJobMappingTransformerToForm(new JobMappingTransformer()),
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

async function updateJobConnection(
  accountId: string,
  job: Job,
  values: SingleTableEditSourceFormValues
): Promise<UpdateJobSourceConnectionResponse> {
  const schema = values.mappings.length > 0 ? values.mappings[0].schema : null;
  const table = values.mappings.length > 0 ? values.mappings[0].table : null;
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${job.id}/source-connection`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(
        new UpdateJobSourceConnectionRequest({
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
          source: new JobSource({
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
          }),
        })
      ),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateJobSourceConnectionResponse.fromJson(await res.json());
}

async function getUpdatedValues(
  accountId: string,
  connectionId: string,
  originalValues: SingleTableEditSourceFormValues,
  mutateConnectionSchemaRes:
    | KeyedMutator<unknown>
    | KeyedMutator<GetConnectionSchemaMapResponse>
): Promise<SingleTableEditSourceFormValues> {
  const [schemaRes, connRes] = await Promise.all([
    getConnectionSchema(accountId, connectionId),
    getConnection(accountId, connectionId),
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

  mutateConnectionSchemaRes(schemaRes);
  return {
    ...originalValues,
    source: {
      fkSourceConnectionId: connectionId,
    },
    mappings,
  };
}
