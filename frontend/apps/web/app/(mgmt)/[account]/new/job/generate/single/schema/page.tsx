'use client';

import { getUniqueSchemasAndTables } from '@/app/(mgmt)/[account]/jobs/[id]/source/components/DataGenConnectionCard';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { getSchemaConstraintHandler } from '@/components/jobs/SchemaTable/SchemaColumns';
import { SchemaTable } from '@/components/jobs/SchemaTable/SchemaTable';
import { setOnboardingConfig } from '@/components/onboarding-checklist/OnboardingChecklist';
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
import { Skeleton } from '@/components/ui/skeleton';
import { useToast } from '@/components/ui/use-toast';
import { useGetAccountOnboardingConfig } from '@/libs/hooks/useGetAccountOnboardingConfig';
import { useGetConnectionForeignConstraints } from '@/libs/hooks/useGetConnectionForeignConstraints';
import { useGetConnectionPrimaryConstraints } from '@/libs/hooks/useGetConnectionPrimaryConstraints';
import { useGetConnectionSchemaMap } from '@/libs/hooks/useGetConnectionSchemaMap';
import { useGetConnectionUniqueConstraints } from '@/libs/hooks/useGetConnectionUniqueConstraints';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { convertMinutesToNanoseconds, getErrorMessage } from '@/util/util';
import {
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
import { ReactElement, useEffect, useMemo, useState } from 'react';
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
const isBrowser = () => typeof window !== 'undefined';

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  const { toast } = useToast();
  const { data: onboardingData, mutate } = useGetAccountOnboardingConfig(
    account?.id ?? ''
  );

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

  const {
    data: connectionSchemaDataMap,
    isLoading: isSchemaDataMapLoading,
    isValidating: isSchemaMapValidating,
  } = useGetConnectionSchemaMap(
    account?.id ?? '',
    connectFormValues.connectionId
  );

  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver<SingleTableSchemaFormValues>(
      SINGLE_TABLE_SCHEMA_FORM_SCHEMA
    ),
    values: schemaFormData,
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

      // updates the onboarding data
      if (!onboardingData?.config?.hasCreatedJob) {
        try {
          await setOnboardingConfig(account.id, {
            hasCreatedSourceConnection:
              onboardingData?.config?.hasCreatedSourceConnection ?? true,
            hasCreatedDestinationConnection:
              onboardingData?.config?.hasCreatedDestinationConnection ?? true,
            hasCreatedJob: true,
            hasInvitedMembers:
              onboardingData?.config?.hasInvitedMembers ?? true,
          });
          mutate();
        } catch (e) {
          toast({
            title: 'Unable to update onboarding status!',
            variant: 'destructive',
          });
        }
      }

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

  const [uniqueSchemas, schemaTableMap] = useMemo(
    () => getUniqueSchemasAndTables(connectionSchemaDataMap?.schemaMap ?? {}),
    [isSchemaMapValidating]
  );

  const { data: primaryConstraints, isValidating: isPkValidating } =
    useGetConnectionPrimaryConstraints(
      account?.id ?? '',
      connectFormValues.connectionId
    );

  const { data: foreignConstraints, isValidating: isFkValidating } =
    useGetConnectionForeignConstraints(
      account?.id ?? '',
      connectFormValues.connectionId
    );

  const { data: uniqueConstraints, isValidating: isUCValidating } =
    useGetConnectionUniqueConstraints(
      account?.id ?? '',
      connectFormValues.connectionId
    );

  const schemaConstraintHandler = useMemo(
    () =>
      getSchemaConstraintHandler(
        connectionSchemaDataMap?.schemaMap ?? {},
        primaryConstraints?.tableConstraints ?? {},
        foreignConstraints?.tableConstraints ?? {},
        uniqueConstraints?.tableConstraints ?? {}
      ),
    [isSchemaMapValidating, isPkValidating, isFkValidating, isUCValidating]
  );

  const selectedSchemaTables = schemaTableMap.get(formValues.schema) ?? [];
  /* this fixes a bug that was happening due to the schema and table values not resetting if you went back and changed
the connection which caused the schema page to not load correctly when you went back to the schema page. This checks if the selected connection contains the schema and
  */
  useEffect(() => {
    const isValidSchemaAndTable =
      form.getValues('schema') != '' && form.getValues('table') != '';
    const tableInSchema = schemaTableMap
      .get(form.getValues('schema'))
      ?.find((item) => item == form.getValues('table'));
    if (isValidSchemaAndTable && !tableInSchema) {
      form.reset({
        ...form.getValues(),
        schema: '',
        table: '',
        mappings: [],
      });
    }
  }, [schemaTableMap, form]);

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
                        {isSchemaDataMapLoading ? (
                          <Skeleton />
                        ) : (
                          Array.from(uniqueSchemas).map((schema) => (
                            <SelectItem
                              className="cursor-pointer"
                              key={schema}
                              value={schema}
                            >
                              {schema}
                            </SelectItem>
                          ))
                        )}
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
                          const dbcols =
                            (connectionSchemaDataMap?.schemaMap ?? {})[
                              `${formValues.schema}.${value}`
                            ] ?? [];
                          form.setValue(
                            'mappings',
                            dbcols.map((s) => {
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
                    type="text"
                    value={field.value
                      .toString()
                      .replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                    onChange={(e) => {
                      const numberValue = parseFloat(
                        e.target.value.replace(/,/g, '')
                      );
                      if (!isNaN(numberValue)) {
                        field.onChange(numberValue);
                      } else {
                        field.onChange(0);
                      }
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
              constraintHandler={schemaConstraintHandler}
              schema={connectionSchemaDataMap?.schemaMap ?? {}}
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

function newDefaultJobMappingTransformerForm(): JobMappingTransformerForm {
  return convertJobMappingTransformerToForm(new JobMappingTransformer({}));
}
