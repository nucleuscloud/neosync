'use client';
import SampleTable from '@/app/(mgmt)/[account]/new/job/aigenerate/single/schema/SampleTable/SampleTable';
import { getAiSampleTableColumns } from '@/app/(mgmt)/[account]/new/job/aigenerate/single/schema/SampleTable/SampleTableColumns';
import SelectModelNames from '@/app/(mgmt)/[account]/new/job/aigenerate/single/schema/SelectModelNames';
import { SampleRecord } from '@/app/(mgmt)/[account]/new/job/aigenerate/single/schema/types';
import { SingleTableEditAiSourceFormValues } from '@/app/(mgmt)/[account]/new/job/job-form-validations';
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
import { getErrorMessage } from '@/util/util';
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  GetConnectionResponse,
  GetConnectionSchemaMapResponse,
  Job,
} from '@neosync/sdk';
import {
  getAiGeneratedData,
  getConnection,
  getConnections,
  getConnectionSchemaMap,
  getConnectionTableConstraints,
  getJob,
  updateJobSourceConnection,
} from '@neosync/sdk/connectquery';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useQueryClient } from '@tanstack/react-query';
import { ColumnDef } from '@tanstack/react-table';
import { ReactElement, useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import {
  getSampleEditAiGeneratedRecordsRequest,
  getSingleTableAiGenerateNumRows,
  getSingleTableAiSchemaTable,
  toSingleTableEditAiGenerateJobSource,
} from '../../../util';
import SchemaPageSkeleton from './SchemaPageSkeleton';

interface Props {
  jobId: string;
}

export default function AiDataGenConnectionCard({
  jobId,
}: Props): ReactElement {
  const { account } = useAccount();

  const {
    data,
    refetch: mutate,
    isLoading: isJobLoading,
  } = useQuery(getJob, { id: jobId }, { enabled: !!jobId });

  const {
    data: connectionsData,
    isLoading: isConnectionsLoading,
    isFetching: isConnectionsValidating,
  } = useQuery(
    getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const connections = connectionsData?.connections ?? [];

  const { mutateAsync: sampleRecords } = useMutation(getAiGeneratedData);

  const form = useForm<SingleTableEditAiSourceFormValues>({
    resolver: yupResolver(SingleTableEditAiSourceFormValues),
    values: getJobSource(data?.job),
    context: { accountId: account?.id },
  });

  const fkSourceConnectionId = form.watch('source.fkSourceConnectionId');

  const {
    data: connectionSchemaDataMap,
    isLoading: isSchemaDataMapLoading,
    isFetching: isSchemaMapValidating,
  } = useQuery(
    getConnectionSchemaMap,
    { connectionId: fkSourceConnectionId },
    { enabled: !!fkSourceConnectionId }
  );
  const { mutateAsync: getConnectionSchemaMapAsync } = useMutation(
    getConnectionSchemaMap
  );
  const queryclient = useQueryClient();

  const { data: tableConstraints, isFetching: isTableConstraintsValidating } =
    useQuery(
      getConnectionTableConstraints,
      { connectionId: fkSourceConnectionId },
      { enabled: !!fkSourceConnectionId }
    );

  const { mutateAsync: updateSourceConnection } = useMutation(
    updateJobSourceConnection
  );
  const { mutateAsync: getConnectionAsync } = useMutation(getConnection);

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
        tdata.push(...tableSchema.schemas);
        cols.push(
          ...getAiSampleTableColumns(
            tableSchema.schemas.map((dbcol) => dbcol.column)
          )
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
    if (!job || !account?.id) {
      return;
    }
    try {
      setIsUpdating(true);
      await updateSourceConnection({
        id: job.id,
        mappings: [],
        source: toSingleTableEditAiGenerateJobSource(values),
      });
      toast.success('Successfully updated job source connection!');
      mutate();
    } catch (err) {
      console.error(err);
      toast.error('Unable to update job source connection', {
        description: getErrorMessage(err),
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
      const output = await sampleRecords(
        getSampleEditAiGeneratedRecordsRequest(form.getValues())
      );
      setaioutput(output.records);
    } catch (err) {
      toast.error('Unable to generate sample data', {
        description: getErrorMessage(err),
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
        value,
        form.getValues(),
        async (id) => {
          const resp = await getConnectionAsync({ id });
          queryclient.setQueryData(
            createConnectQueryKey({
              schema: getConnection,
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
              schema: getConnectionSchemaMap,
              input: { connectionId: id },
              cardinality: undefined,
            }),
            resp
          );
          return resp;
        }
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
      toast.error(
        'Unable to get connection schema on table constraint source change.',
        {
          description: getErrorMessage(err),
        }
      );
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
        <FormField
          control={form.control}
          name="schema.generateBatchSize"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Batch Size</FormLabel>
              <FormDescription>
                The batch size used when asking the model to generate records.
                Useful for large datasets or prompts that may exceed AI token
                limits. Smaller is generally better.
              </FormDescription>
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
        numRows: 1,
        generateBatchSize: 1,
        userPrompt: '',
      },
    };
  }
  let numRows = 0;
  let genBatchSize = 10;
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
    numRows = getSingleTableAiGenerateNumRows(job.source.options.config.value);

    if (job.source.options.config.value.generateBatchSize) {
      genBatchSize = Number(job.source.options.config.value.generateBatchSize);
    } else {
      // batch size has not been set by the user. Set it to our default of 10, or num rows, whichever is lower
      genBatchSize = Math.min(genBatchSize, numRows);
    }
    const schematable = getSingleTableAiSchemaTable(
      job.source.options.config.value
    );
    schema = schematable.schema;
    table = schematable.table;
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
      generateBatchSize: genBatchSize,
    },
  };
}

async function getUpdatedValues(
  connectionId: string,
  originalValues: SingleTableEditAiSourceFormValues,
  getConnectionById: (id: string) => Promise<GetConnectionResponse>,
  getConnectionSchemaMapAsync: (
    id: string
  ) => Promise<GetConnectionSchemaMapResponse>
): Promise<SingleTableEditAiSourceFormValues> {
  const [schemaRes, connRes] = await Promise.all([
    getConnectionSchemaMapAsync(connectionId),
    getConnectionById(connectionId),
  ]);

  if (!schemaRes || !connRes) {
    return originalValues;
  }

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
