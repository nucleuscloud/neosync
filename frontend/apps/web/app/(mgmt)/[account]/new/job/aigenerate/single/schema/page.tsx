'use client';

import ButtonText from '@/components/ButtonText';
import { Action } from '@/components/DualListBox/DualListBox';
import Spinner from '@/components/Spinner';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import {
  AiSchemaTable,
  AiSchemaTableRecord,
} from '@/components/jobs/SchemaTable/AiSchemaTable';
import { getSchemaConstraintHandler } from '@/components/jobs/SchemaTable/schema-constraint-handler';
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
import { Textarea } from '@/components/ui/textarea';
import { useToast } from '@/components/ui/use-toast';
import { useGetAccountOnboardingConfig } from '@/libs/hooks/useGetAccountOnboardingConfig';
import { useGetConnectionForeignConstraints } from '@/libs/hooks/useGetConnectionForeignConstraints';
import { useGetConnectionPrimaryConstraints } from '@/libs/hooks/useGetConnectionPrimaryConstraints';
import { useGetConnectionSchemaMap } from '@/libs/hooks/useGetConnectionSchemaMap';
import { useGetConnectionUniqueConstraints } from '@/libs/hooks/useGetConnectionUniqueConstraints';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { convertMinutesToNanoseconds, getErrorMessage } from '@/util/util';
import { toJobDestinationOptions } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  ActivityOptions,
  AiGenerateSourceOptions,
  AiGenerateSourceSchemaOption,
  AiGenerateSourceTableOption,
  Connection,
  CreateJobRequest,
  CreateJobResponse,
  DatabaseTable,
  GetAccountOnboardingConfigResponse,
  GetAiGeneratedDataRequest,
  JobDestination,
  JobSource,
  JobSourceOptions,
  RetryPolicy,
  WorkflowOptions,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ColumnDef } from '@tanstack/react-table';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
// import useFormPersist from 'react-hook-form-persist';
import FormPersist from '@/app/(mgmt)/FormPersist';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps, {
  getJobProgressSteps,
} from '../../../JobsProgressSteps';
import {
  DefineFormValues,
  SingleTableAiConnectFormValues,
  SingleTableAiSchemaFormValues,
} from '../../../schema';
import SampleTable from './SampleTable/SampleTable';
import { getAiSampleTableColumns } from './SampleTable/SampleTableColumns';
import SelectModelNames from './SelectModelNames';
import { SampleRecord } from './types';

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  const { toast } = useToast();
  const { data: onboardingData, mutate } = useGetAccountOnboardingConfig(
    account?.id ?? ''
  );
  const [aioutput, setaioutput] = useState<SampleRecord[]>([]);

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
  const connectFormKey = `${sessionPrefix}-new-job-single-table-ai-connect`;
  const [connectFormValues] = useSessionStorage<SingleTableAiConnectFormValues>(
    connectFormKey,
    {
      sourceId: '',
      fkSourceConnectionId: '',
      destination: {
        connectionId: '',
        destinationOptions: {},
      },
    }
  );

  const {
    data: connectionSchemaDataMap,
    isLoading: isSchemaMapLoading,
    isValidating: isSchemaMapValidating,
  } = useGetConnectionSchemaMap(
    account?.id ?? '',
    connectFormValues.fkSourceConnectionId
  );

  const formKey = `${sessionPrefix}-new-job-single-table-ai-schema`;

  const [schemaFormData] = useSessionStorage<SingleTableAiSchemaFormValues>(
    formKey,
    {
      numRows: 10,
      model: 'gpt-3.5-turbo',
      userPrompt: '',
      schema: '',
      table: '',
    }
  );

  const form = useForm<SingleTableAiSchemaFormValues>({
    resolver: yupResolver<SingleTableAiSchemaFormValues>(
      SingleTableAiSchemaFormValues
    ),
    values: schemaFormData,
  });

  const [isClient, setIsClient] = useState(false);
  useEffect(() => {
    setIsClient(true);
  }, []);

  const [isSampling, setIsSampling] = useState(false);

  async function onSubmit(values: SingleTableAiSchemaFormValues) {
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
          const resp = await setOnboardingConfig(account.id, {
            hasCreatedSourceConnection:
              onboardingData?.config?.hasCreatedSourceConnection ?? true,
            hasCreatedDestinationConnection:
              onboardingData?.config?.hasCreatedDestinationConnection ?? true,
            hasCreatedJob: true,
            hasInvitedMembers:
              onboardingData?.config?.hasInvitedMembers ?? true,
          });
          mutate(
            new GetAccountOnboardingConfigResponse({
              config: resp.config,
            })
          );
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

  const { data: primaryConstraints, isValidating: isPkValidating } =
    useGetConnectionPrimaryConstraints(
      account?.id ?? '',
      connectFormValues.fkSourceConnectionId
    );

  const { data: foreignConstraints, isValidating: isFkValidating } =
    useGetConnectionForeignConstraints(
      account?.id ?? '',
      connectFormValues.fkSourceConnectionId
    );

  const { data: uniqueConstraints, isValidating: isUCValidating } =
    useGetConnectionUniqueConstraints(
      account?.id ?? '',
      connectFormValues.fkSourceConnectionId
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
  const [selectedTables, setSelectedTables] = useState<Set<string>>(new Set());

  useEffect(() => {
    if (isSchemaMapLoading || selectedTables.size > 0) {
      return;
    }
    const js = schemaFormData;
    if (js.schema && js.table) {
      onSelectedTableToggle(new Set([`${js.schema}.${js.table}`]), 'add');
    }
  }, [isSchemaMapLoading]);

  async function onSampleClick(): Promise<void> {
    if (!account?.id || isSampling) {
      return;
    }
    try {
      setIsSampling(true);
      const output = await sample(
        connectFormValues,
        form.getValues(),
        account.id
      );
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
      form.setValue('schema', '');
      form.setValue('table', '');
      return;
    }
    if (action === 'add' || action === 'add-all') {
      setSelectedTables(items);
      const item = Array.from(items)[0];
      const [schema, table] = item.split('.');
      form.setValue('schema', schema);
      form.setValue('table', table);
    } else {
      setSelectedTables(new Set());
      form.setValue('schema', '');
      form.setValue('table', '');
    }
  }

  const [formSchema, formTable] = form.watch(['schema', 'table']);

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

  return (
    <div className="flex flex-col gap-5">
      <FormPersist formKey={formKey} form={form} />
      <OverviewContainer
        Header={
          <PageHeader
            header="Schema"
            progressSteps={
              <JobsProgressSteps
                steps={getJobProgressSteps('ai-generate-table')}
                stepName={'schema'}
              />
            }
          />
        }
        containerClassName="connect-page"
      >
        <div />
      </OverviewContainer>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          {isClient && (
            <AiSchemaTable
              data={tableData}
              constraintHandler={schemaConstraintHandler}
              schema={connectionSchemaDataMap?.schemaMap ?? {}}
              isSchemaDataReloading={isSchemaMapValidating}
              selectedTables={selectedTables}
              onSelectedTableToggle={onSelectedTableToggle}
            />
          )}

          <FormField
            control={form.control}
            name="model"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Model Name</FormLabel>
                <div className="flex flex-col md:flex-row gap-2 md:items-center">
                  <FormDescription>
                    The name of the model to use.
                  </FormDescription>
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
            name="userPrompt"
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
            name="numRows"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Number of Rows</FormLabel>
                <FormDescription>
                  The number of rows to generate.
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

          <div className="flex flex-row gap-1 justify-between">
            <Button key="back" type="button" onClick={() => router.back()}>
              Back
            </Button>
            <div className="flex flex-row gap-4">
              <Button type="button" onClick={() => onSampleClick()}>
                <ButtonText
                  leftIcon={isSampling ? <Spinner /> : undefined}
                  text="Sample"
                />
              </Button>
              <Button key="submit" type="submit">
                Submit
              </Button>
            </div>
          </div>
        </form>
      </Form>
    </div>
  );
}

async function sample(
  connect: SingleTableAiConnectFormValues,
  schemaform: SingleTableAiSchemaFormValues,
  accountId: string
): Promise<SampleRecord[]> {
  const body = new GetAiGeneratedDataRequest({
    aiConnectionId: connect.sourceId,
    count: BigInt(10),
    modelName: schemaform.model,
    userPrompt: schemaform.userPrompt,
    dataConnectionId: connect.fkSourceConnectionId,
    table: new DatabaseTable({
      schema: schemaform.schema,
      table: schemaform.table,
    }),
  });

  const res = await fetch(
    `/api/accounts/${accountId}/connections/${connect.sourceId}/generate`,
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

async function createNewJob(
  define: DefineFormValues,
  connect: SingleTableAiConnectFormValues,
  schemaForm: SingleTableAiSchemaFormValues,
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
    mappings: [],
    source: new JobSource({
      options: new JobSourceOptions({
        config: {
          case: 'aiGenerate',
          value: new AiGenerateSourceOptions({
            aiConnectionId: connect.sourceId,
            modelName: schemaForm.model,
            fkSourceConnectionId: connect.fkSourceConnectionId,
            userPrompt: schemaForm.userPrompt,
            schemas: [
              new AiGenerateSourceSchemaOption({
                schema: schemaForm.schema,
                tables: [
                  new AiGenerateSourceTableOption({
                    table: schemaForm.table,
                    rowCount: BigInt(schemaForm.numRows),
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
        connectionId: connect.destination.connectionId,
        options: toJobDestinationOptions(
          connect.destination,
          connectionIdMap.get(connect.destination.connectionId)
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
