'use client';

import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { SchemaTable } from '@/components/jobs/SchemaTable/SchemaTable';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
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
import { useGetConnectionForeignConstraints } from '@/libs/hooks/useGetConnectionForeignConstraints';
import { useGetConnectionPrimaryConstraints } from '@/libs/hooks/useGetConnectionPrimaryConstraints';
import { useGetConnectionSchema } from '@/libs/hooks/useGetConnectionSchema';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { convertMinutesToNanoseconds, getErrorMessage } from '@/util/util';
import {
  JobMappingFormValues,
  JobMappingTransformerForm,
  convertJobMappingTransformerFormToJobMappingTransformer,
  convertJobMappingTransformerToForm,
  toJobDestinationOptions,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  ActivityOptions,
  Connection,
  CreateJobRequest,
  CreateJobResponse,
  DatabaseColumn,
  GenerateSourceOptions,
  GenerateSourceSchemaOption,
  GenerateSourceTableOption,
  JobDestination,
  JobMapping,
  JobMappingTransformer,
  JobSource,
  JobSourceOptions,
  RetryPolicy,
  WorkflowOptions,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps, { DATA_GEN_STEPS } from '../../../JobsProgressSteps';
import {
  DefineFormValues,
  SINGLE_TABLE_SCHEMA_FORM_SCHEMA,
  SingleTableConnectFormValues,
  SingleTableSchemaFormValues,
} from '../../../schema';
import { ColumnMetadata } from '../../../schema/page';
const isBrowser = () => typeof window !== 'undefined';

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  const { toast } = useToast();

  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/${account?.name}/new/job`);
    }
  }, [searchParams?.sessionId]);
  const { data: connectionsData } = useGetConnections(account?.id ?? '');
  const connections = connectionsData?.connections ?? [];

  const sessionPrefix = searchParams?.sessionId ?? '';

  // Used to complete the whole form
  const defineFormKey = `${sessionPrefix}-new-job-define`;
  const [defineFormValues] = useSessionStorage<DefineFormValues>(
    defineFormKey,
    { jobName: '' }
  );

  const connectFormKey = `${sessionPrefix}-new-job-single-table-connect`;
  const [connectFormValues] = useSessionStorage<SingleTableConnectFormValues>(
    connectFormKey,
    {
      connectionId: '',
      destinationOptions: {},
    }
  );
  const { data: connSchemaData } = useGetConnectionSchema(
    account?.id ?? '',
    connectFormValues.connectionId
  );

  const formKey = `${sessionPrefix}-new-job-single-table-schema`;

  const [schemaFormData] = useSessionStorage<SingleTableSchemaFormValues>(
    formKey,
    {
      mappings: [],
      numRows: 10,
      schema: '',
      table: '',
    }
  );

  const { data: connectionSchemaData } = useGetConnectionSchema(
    account?.id ?? '',
    connectFormValues.connectionId
  );

  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver<SingleTableSchemaFormValues>(
      SINGLE_TABLE_SCHEMA_FORM_SCHEMA
    ),
    values: getFormValues(connectionSchemaData?.schemas ?? [], schemaFormData),
  });

  useFormPersist(formKey, {
    watch: form.watch,
    setValue: form.setValue,
    storage: isBrowser() ? window.sessionStorage : undefined,
  });
  const [isClient, setIsClient] = useState(false);
  useEffect(() => {
    setIsClient(true);
  }, []);

  async function onSubmit(values: SingleTableSchemaFormValues) {
    if (!account) {
      return;
    }
    try {
      const job = await createNewJob(
        defineFormValues,
        connectFormValues,
        values,
        account.id,
        connections
      );
      toast({
        title: 'Successfully created job!',
        variant: 'success',
      });
      window.sessionStorage.removeItem(defineFormKey);
      window.sessionStorage.removeItem(connectFormKey);
      window.sessionStorage.removeItem(formKey);
      if (job.job?.id) {
        router.push(`/${account?.name}/jobs/${job.job.id}`);
      } else {
        router.push(`/${account?.name}/jobs`);
      }
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to create job',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  const formValues = form.watch();
  const schemaTableData = formValues.mappings ?? [];

  const uniqueSchemas = Array.from(
    new Set(connSchemaData?.schemas.map((s) => s.schema))
  );
  const schemaTableMap = getSchemaTableMap(connSchemaData?.schemas ?? []);

  const { data: primaryConstraints } = useGetConnectionPrimaryConstraints(
    account?.id ?? '',
    connectFormValues.connectionId
  );

  const { data: foreignConstraints } = useGetConnectionForeignConstraints(
    account?.id ?? '',
    connectFormValues.connectionId
  );

  const columnMetadata: ColumnMetadata = {
    pk: primaryConstraints?.tableConstraints ?? {},
    fk: foreignConstraints?.tableConstraints ?? {},
    isNullable: connSchemaData?.schemas ?? [],
  };

  const selectedSchemaTables = schemaTableMap.get(formValues.schema) ?? [];

  return (
    <div className="flex flex-col gap-5">
      <OverviewContainer
        Header={
          <PageHeader
            header="Schema"
            progressSteps={
              <JobsProgressSteps steps={DATA_GEN_STEPS} stepName={'schema'} />
            }
          />
        }
        containerClassName="connect-page"
      >
        <div />
      </OverviewContainer>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <FormField
            control={form.control}
            name="schema"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Schema</FormLabel>
                <FormDescription>The name of the schema.</FormDescription>
                <FormControl>
                  {isClient && (
                    <Select
                      onValueChange={(value: string) => {
                        if (value) {
                          field.onChange(value);
                          form.setValue('table', ''); // reset the table value because it may no longer apply
                        }
                      }}
                      value={field.value}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select a schema..." />
                      </SelectTrigger>
                      <SelectContent>
                        {uniqueSchemas.map((schema) => (
                          <SelectItem
                            className="cursor-pointer"
                            key={schema}
                            value={schema}
                          >
                            {schema}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="table"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Table Name</FormLabel>
                <FormDescription>The name of the table.</FormDescription>
                <FormControl>
                  {isClient && (
                    <Select
                      disabled={!formValues.schema}
                      onValueChange={(value: string) => {
                        if (value) {
                          field.onChange(value);
                          form.setValue(
                            'mappings',
                            (connectionSchemaData?.schemas ?? [])
                              .filter(
                                (s) =>
                                  s.schema === formValues.schema &&
                                  s.table === value
                              )
                              .map((s) => {
                                return {
                                  schema: s.schema,
                                  table: s.table,
                                  column: s.column,
                                  dataType: s.dataType,
                                  transformer:
                                    newDefaultJobMappingTransformerForm(),
                                  isNullable: s.isNullable,
                                };
                              })
                          );
                        }
                      }}
                      value={field.value}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select a table..." />
                      </SelectTrigger>
                      <SelectContent>
                        {selectedSchemaTables.map((table) => (
                          <SelectItem
                            className="cursor-pointer"
                            key={table}
                            value={table}
                          >
                            {table}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
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
                <FormDescription>
                  The number of rows to generate.
                </FormDescription>
                <FormControl>
                  <Input
                    type="number"
                    {...field}
                    onChange={(e) => {
                      field.onChange(e.target.valueAsNumber);
                    }}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          {isClient && formValues.schema && formValues.table && (
            <SchemaTable
              data={schemaTableData}
              excludeInputReqTransformers
              columnMetadata={columnMetadata}
              jobType={'generate'}
            />
          )}
          {form.formState.errors.root && (
            <Alert variant="destructive">
              <AlertTitle className="flex flex-row space-x-2 justify-center">
                <ExclamationTriangleIcon />
                <p>Please fix form errors and try again.</p>
              </AlertTitle>
            </Alert>
          )}
          <div className="flex flex-row gap-1 justify-between">
            <Button key="back" type="button" onClick={() => router.back()}>
              Back
            </Button>
            <Button key="submit" type="submit">
              Submit
            </Button>
          </div>
        </form>
      </Form>
    </div>
  );
}

async function createNewJob(
  define: DefineFormValues,
  connect: SingleTableConnectFormValues,
  schema: SingleTableSchemaFormValues,
  accountId: string,
  connections: Connection[]
): Promise<CreateJobResponse> {
  const connectionIdMap = new Map(
    connections.map((connection) => [connection.id, connection])
  );
  let workflowOptions: WorkflowOptions | undefined = undefined;
  if (define.workflowSettings?.runTimeout) {
    workflowOptions = new WorkflowOptions({
      runTimeout: convertMinutesToNanoseconds(
        define.workflowSettings.runTimeout
      ),
    });
  }
  let syncOptions: ActivityOptions | undefined = undefined;
  if (define.syncActivityOptions) {
    const formSyncOpts = define.syncActivityOptions;
    syncOptions = new ActivityOptions({
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
  const body = new CreateJobRequest({
    accountId,
    jobName: define.jobName,
    cronSchedule: define.cronSchedule,
    initiateJobRun: define.initiateJobRun,
    mappings: schema.mappings.map((m) => {
      return new JobMapping({
        schema: m.schema,
        table: m.table,
        column: m.column,
        transformer: convertJobMappingTransformerFormToJobMappingTransformer(
          m.transformer
        ),
      });
    }),
    source: new JobSource({
      options: new JobSourceOptions({
        config: {
          case: 'generate',
          value: new GenerateSourceOptions({
            fkSourceConnectionId: connect.connectionId,
            schemas: [
              new GenerateSourceSchemaOption({
                schema: schema.schema,
                tables: [
                  new GenerateSourceTableOption({
                    rowCount: BigInt(schema.numRows),
                    table: schema.table,
                  }),
                ],
              }),
            ],
          }),
        },
      }),
    }),
    destinations: [
      new JobDestination({
        connectionId: connect.connectionId,
        options: toJobDestinationOptions(
          connect,
          connectionIdMap.get(connect.connectionId)
        ),
      }),
    ],
    workflowOptions: workflowOptions,
    syncOptions: syncOptions,
  });

  const res = await fetch(`/api/accounts/${accountId}/jobs`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CreateJobResponse.fromJson(await res.json());
}

function getSchemaTableMap(schemas: DatabaseColumn[]): Map<string, string[]> {
  const map = new Map<string, Set<string>>();
  schemas.forEach((schema) => {
    const set = map.get(schema.schema);
    if (set) {
      set.add(schema.table);
    } else {
      map.set(schema.schema, new Set([schema.table]));
    }
  });

  const outMap = new Map<string, string[]>();
  map.forEach((tableSet, schema) => outMap.set(schema, Array.from(tableSet)));
  return outMap;
}

function getFormValues(
  dbCols: DatabaseColumn[],
  existingData: SingleTableSchemaFormValues | undefined
): SingleTableSchemaFormValues {
  const schema = existingData?.schema ?? '';
  const table = existingData?.table ?? '';
  const defaultMappings = dbCols
    .filter((dbCol) => dbCol.schema === schema && dbCol.table === table)
    .map((dbCol) => {
      return {
        ...dbCol,
        transformer: newDefaultJobMappingTransformerForm(),
      };
    });
  const existingMappings = (existingData?.mappings ?? []).filter(
    (mapping) => mapping.schema === schema && mapping.table === table
  );
  const mappingMap = new Map<string, JobMappingFormValues>();
  defaultMappings.forEach((mapping) =>
    mappingMap.set(
      `${mapping.schema}-${mapping.table}-${mapping.column}`,
      mapping
    )
  );
  existingMappings.forEach((mapping) =>
    mappingMap.set(
      `${mapping.schema}-${mapping.table}-${mapping.column}`,
      mapping
    )
  );
  return {
    numRows: existingData?.numRows ?? 10,
    schema,
    table,
    mappings: Array.from(mappingMap.values()),
  };
}

function newDefaultJobMappingTransformerForm(): JobMappingTransformerForm {
  return convertJobMappingTransformerToForm(new JobMappingTransformer({}));
}
