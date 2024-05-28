'use client';
import SampleTable from '@/app/(mgmt)/[account]/new/job/aigenerate/single/schema/SampleTable/SampleTable';
import { getAiSampleTableColumns } from '@/app/(mgmt)/[account]/new/job/aigenerate/single/schema/SampleTable/SampleTableColumns';
import SelectModelNames from '@/app/(mgmt)/[account]/new/job/aigenerate/single/schema/SelectModelNames';
import { SampleRecord } from '@/app/(mgmt)/[account]/new/job/aigenerate/single/schema/types';
import { SingleTableEditAiSourceFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import ButtonText from '@/components/ButtonText';
import { Action } from '@/components/DualListBox/DualListBox';
import Spinner from '@/components/Spinner';
import {
  AiSchemaTable,
  AiSchemaTableRecord,
} from '@/components/jobs/SchemaTable/AiSchemaTable';
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
import { Textarea } from '@/components/ui/textarea';
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
import { getErrorMessage } from '@/util/util';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  AiGenerateSourceOptions,
  AiGenerateSourceSchemaOption,
  DatabaseTable,
  GenerateSourceTableOption,
  GetAiGeneratedDataRequest,
  Job,
  JobSource,
  JobSourceOptions,
  UpdateJobSourceConnectionRequest,
  UpdateJobSourceConnectionResponse,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ColumnDef } from '@tanstack/react-table';
import { ReactElement, useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { KeyedMutator } from 'swr';
import SchemaPageSkeleton from './SchemaPageSkeleton';

interface Props {
  jobId: string;
}

export default function AiDataGenConnectionCard({
  jobId,
}: Props): ReactElement {
  const { toast } = useToast();
  const { account } = useAccount();

  const {
    data,
    mutate,
    isLoading: isJobLoading,
  } = useGetJob(account?.id ?? '', jobId);

  const {
    isLoading: isConnectionsLoading,
    isValidating: isConnectionsValidating,
    data: connectionsData,
  } = useGetConnections(account?.id ?? '');
  const connections = connectionsData?.connections ?? [];

  const form = useForm<SingleTableEditAiSourceFormValues>({
    resolver: yupResolver(SingleTableEditAiSourceFormValues),
    values: getJobSource(data?.job),
    context: { accountId: account?.id },
  });

  const fkSourceConnectionId = form.watch('source.fkSourceConnectionId');

  const {
    data: connectionSchemaDataMap,
    isLoading: isSchemaDataMapLoading,
    isValidating: isSchemaMapValidating,
    mutate: mutateGetConnectionSchemaMap,
  } = useGetConnectionSchemaMap(account?.id ?? '', fkSourceConnectionId);

  const { data: tableConstraints, isValidating: isTableConstraintsValidating } =
    useGetConnectionTableConstraints(account?.id ?? '', fkSourceConnectionId);

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

  const [aioutput, setaioutput] = useState<SampleRecord[]>([]);
  const [isSampling, setIsSampling] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);

  const [selectedTables, setSelectedTables] = useState<Set<string>>(new Set());

  const connectionsMap = useMemo(
    () => new Map(connections.map((c) => [c.id, c])),
    [isConnectionsValidating]
  );

  useEffect(() => {
    if (
      isJobLoading ||
      data?.job?.source?.options?.config.case !== 'aiGenerate'
    ) {
      return;
    }
    const js = getJobSource(data.job);
    if (js.schema.schema && js.schema.table) {
      onSelectedTableToggle(
        new Set([`${js.schema.schema}.${js.schema.table}`]),
        'add'
      );
    }
  }, [isJobLoading]);

  const [formSchema, formTable, formSourceId] = form.watch([
    'schema.schema',
    'schema.table',
    'source.sourceId',
  ]);

  const [tableData, columns] = useMemo(() => {
    const tdata: AiSchemaTableRecord[] = [];
    const cols: ColumnDef<SampleRecord>[] = [];
    if (formSchema && formTable && connectionSchemaDataMap?.schemaMap) {
      const tableSchema =
        connectionSchemaDataMap.schemaMap[`${formSchema}.${formTable}`];
      if (tableSchema) {
        tdata.push(...tableSchema);
        cols.push(
          ...getAiSampleTableColumns(tableSchema.map((dbcol) => dbcol.column))
        );
      }
    }
    return [tdata, cols];
  }, [formSchema, formTable, isSchemaMapValidating]);

  const sourceConn = connectionsMap.get(formSourceId);

  if (isJobLoading || isSchemaDataMapLoading || isConnectionsLoading) {
    return <SchemaPageSkeleton />;
  }

  async function onSubmit(values: SingleTableEditAiSourceFormValues) {
    const job = data?.job;
    if (!job) {
      return;
    }
    try {
      setIsUpdating(true);
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
    } finally {
      setIsUpdating(false);
    }
  }

  async function onSampleClick(): Promise<void> {
    if (!account?.id || isSampling || !data?.job) {
      return;
    }
    try {
      setIsSampling(true);
      const output = await sample(form.getValues(), account.id);
      setaioutput(output);
    } catch (err) {
      toast({
        title: 'Unable to generate sample data',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    } finally {
      setIsSampling(false);
    }
  }

  function onSelectedTableToggle(items: Set<string>, action: Action): void {
    if (items.size === 0) {
      setSelectedTables(new Set());
      form.setValue('schema.schema', '');
      form.setValue('schema.table', '');
      return;
    }
    if (action === 'add') {
      setSelectedTables(items);
      const item = Array.from(items)[0];
      const [schema, table] = item.split('.');
      form.setValue('schema.schema', schema);
      form.setValue('schema.table', table);
    } else {
      setSelectedTables(new Set());
      form.setValue('schema.schema', '');
      form.setValue('schema.table', '');
    }
  }

  async function onTableConstraintSourceChange(value: string): Promise<void> {
    try {
      const newValues = await getUpdatedValues(
        account?.id ?? '',
        value,
        form.getValues(),
        mutateGetConnectionSchemaMap
      );
      form.reset(newValues);
      if (newValues.schema.schema && newValues.schema.table) {
        setSelectedTables(
          new Set([`${newValues.schema.schema}.${newValues.schema.table}`])
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

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <FormField
          control={form.control}
          name="source.sourceId"
          render={({ field }) => (
            <FormItem>
              <FormLabel>OpenAI Connection Source</FormLabel>
              <FormDescription>
                The OpenAI SDK connection in use.
              </FormDescription>
              <FormControl>
                <Select
                  value={field.value}
                  onValueChange={async (value) => {
                    if (!value) {
                      return;
                    }
                    field.onChange(value);
                  }}
                >
                  <SelectTrigger>
                    <SelectValue placeholder={sourceConn?.name} />
                  </SelectTrigger>
                  <SelectContent>
                    {connections
                      .filter(
                        (c) =>
                          c.connectionConfig?.config.case === 'openaiConfig'
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
                    <SelectValue placeholder={sourceConn?.name} />
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

        <AiSchemaTable
          data={tableData}
          constraintHandler={schemaConstraintHandler}
          schema={connectionSchemaDataMap?.schemaMap ?? {}}
          isSchemaDataReloading={isSchemaMapValidating}
          selectedTables={selectedTables}
          onSelectedTableToggle={onSelectedTableToggle}
        />

        <FormField
          control={form.control}
          name="schema.model"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Model Name</FormLabel>
              <div className="flex flex-col md:flex-row gap-2 md:items-center">
                <FormDescription>The name of the model to use.</FormDescription>
                <div>
                  <SelectModelNames
                    onSelected={(modelName) => field.onChange(modelName)}
                  />
                </div>
              </div>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="schema.userPrompt"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Prompt</FormLabel>
              <FormDescription>
                Optionally provide a prompt to give further context to the
                model. This is highly recommended! The table schema will be
                appended to the end of this prompt automatically.
              </FormDescription>
              <FormControl>
                <Textarea
                  {...field}
                  autoComplete="off"
                  autoCapitalize="off"
                  autoCorrect="off"
                  spellCheck={false}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="schema.numRows"
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

        {form.formState.errors.root && (
          <Alert variant="destructive">
            <AlertTitle className="flex flex-row space-x-2 justify-center">
              <ExclamationTriangleIcon />
              <p>Please fix form errors and try again.</p>
            </AlertTitle>
          </Alert>
        )}
        <SampleTable columns={columns} records={aioutput} />

        <div className="flex flex-row gap-2 justify-end">
          <Button
            variant="secondary"
            type="button"
            onClick={() => onSampleClick()}
          >
            <ButtonText
              leftIcon={isSampling ? <Spinner /> : undefined}
              text="Sample"
            />
          </Button>
          <Button type="submit">
            <ButtonText
              leftIcon={isUpdating ? <Spinner /> : undefined}
              text="Update"
            />
          </Button>
        </div>
      </form>
    </Form>
  );
}

function getJobSource(job?: Job): SingleTableEditAiSourceFormValues {
  if (!job) {
    return {
      source: {
        sourceId: '',
        fkSourceConnectionId: '',
      },
      schema: {
        model: '',
        schema: '',
        table: '',
        numRows: 0,
        userPrompt: '',
      },
    };
  }
  let numRows = 0;
  let schema = '';
  let table = '';
  let model = '';
  let userPrompt = '';
  let sourceId = '';
  let fkSourceConnectionId = '';
  if (job.source?.options?.config.case === 'aiGenerate') {
    sourceId = job.source.options.config.value.aiConnectionId;
    fkSourceConnectionId =
      job.source.options.config.value.fkSourceConnectionId ?? '';
    model = job.source.options.config.value.modelName;
    userPrompt = job.source.options.config.value.userPrompt ?? '';
    const srcSchemas = job.source.options.config.value.schemas;
    if (srcSchemas.length > 0) {
      const tables = srcSchemas[0].tables;
      if (tables.length > 0) {
        numRows = Number(tables[0].rowCount); // this will be an issue if the number is bigger than what js allows
        schema = srcSchemas[0].schema;
        table = tables[0].table;
      }
    }
  }

  return {
    source: {
      sourceId,
      fkSourceConnectionId,
    },
    schema: {
      numRows,
      schema,
      table,
      model,
      userPrompt,
    },
  };
}

async function sample(
  schemaform: SingleTableEditAiSourceFormValues,
  accountId: string
): Promise<SampleRecord[]> {
  const body = new GetAiGeneratedDataRequest({
    aiConnectionId: schemaform.source.sourceId,
    count: BigInt(10),
    modelName: schemaform.schema.model,
    userPrompt: schemaform.schema.userPrompt,
    dataConnectionId: schemaform.source.fkSourceConnectionId,
    table: new DatabaseTable({
      schema: schemaform.schema.schema,
      table: schemaform.schema.table,
    }),
  });

  const res = await fetch(
    `/api/accounts/${accountId}/connections/${schemaform.source.sourceId}/generate`,
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

async function updateJobConnection(
  accountId: string,
  job: Job,
  values: SingleTableEditAiSourceFormValues
): Promise<UpdateJobSourceConnectionResponse> {
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
          mappings: [],
          source: new JobSource({
            options: new JobSourceOptions({
              config: {
                case: 'aiGenerate',
                value: new AiGenerateSourceOptions({
                  aiConnectionId: values.source.sourceId,
                  fkSourceConnectionId: values.source.fkSourceConnectionId,
                  modelName: values.schema.model,
                  userPrompt: values.schema.userPrompt,
                  schemas: [
                    new AiGenerateSourceSchemaOption({
                      schema: values.schema.schema,
                      tables: [
                        new GenerateSourceTableOption({
                          table: values.schema.table,
                          rowCount: BigInt(values.schema.numRows),
                        }),
                      ],
                    }),
                  ],
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
  originalValues: SingleTableEditAiSourceFormValues,
  mutateConnectionSchemaRes:
    | KeyedMutator<unknown>
    | KeyedMutator<GetConnectionSchemaMapResponse>
): Promise<SingleTableEditAiSourceFormValues> {
  const [schemaRes, connRes] = await Promise.all([
    getConnectionSchema(accountId, connectionId),
    getConnection(accountId, connectionId),
  ]);

  if (!schemaRes || !connRes) {
    return originalValues;
  }

  mutateConnectionSchemaRes(schemaRes);
  let schema = originalValues.schema.schema;
  let table = originalValues.schema.table;
  if (
    !schemaRes.schemaMap[
      `${originalValues.schema.schema}.${originalValues.schema.table}`
    ]
  ) {
    schema = '';
    table = '';
  }
  return {
    source: {
      sourceId: originalValues.source.sourceId,
      fkSourceConnectionId: connectionId,
    },
    schema: {
      ...originalValues.schema,
      schema,
      table,
    },
  };
}
