'use client';

import FormPersist from '@/app/(mgmt)/FormPersist';
import {
  clearNewJobSession,
  fromStructToRecord,
  getCreateNewSingleTableAiGenerateJobRequest,
  getNewJobSessionKeys,
  getSampleAiGeneratedRecordsRequest,
} from '@/app/(mgmt)/[account]/jobs/util';
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
import { getSingleOrUndefined } from '@/libs/utils';
import { getErrorMessage } from '@/util/util';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  createJob,
  getAiGeneratedData,
  getConnections,
  getConnectionSchemaMap,
  getConnectionTableConstraints,
} from '@neosync/sdk/connectquery';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ColumnDef } from '@tanstack/react-table';
import { useRouter } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps, {
  getJobProgressSteps,
} from '../../../JobsProgressSteps';
import {
  DefineFormValues,
  SingleTableAiConnectFormValues,
  SingleTableAiSchemaFormValues,
} from '../../../job-form-validations';
import SampleTable from './SampleTable/SampleTable';
import { getAiSampleTableColumns } from './SampleTable/SampleTableColumns';
import SelectModelNames from './SelectModelNames';
import { SampleRecord } from './types';

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  const posthog = usePostHog();
  const [aioutput, setaioutput] = useState<SampleRecord[]>([]);

  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/${account?.name}/new/job`);
    }
  }, [searchParams?.sessionId]);
  const { data: connectionsData } = useQuery(
    getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const connections = connectionsData?.connections ?? [];

  const { mutateAsync: createJobAsync } = useMutation(createJob);
  const { mutateAsync: sampleRecords } = useMutation(getAiGeneratedData);

  const sessionPrefix = getSingleOrUndefined(searchParams?.sessionId) ?? '';
  const sessionKeys = getNewJobSessionKeys(sessionPrefix);

  // Used to complete the whole form
  const defineFormKey = sessionKeys.global.define;
  const [defineFormValues] = useSessionStorage<DefineFormValues>(
    defineFormKey,
    { jobName: '' }
  );
  const connectFormKey = sessionKeys.aigenerate.connect;
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
    isFetching: isSchemaMapValidating,
  } = useQuery(
    getConnectionSchemaMap,
    { connectionId: connectFormValues.fkSourceConnectionId },
    { enabled: !!connectFormValues.fkSourceConnectionId }
  );

  const formKey = sessionKeys.aigenerate.schema;
  const [schemaFormData] = useSessionStorage<SingleTableAiSchemaFormValues>(
    formKey,
    {
      numRows: 10,
      generateBatchSize: 10,
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
      const connMap = new Map(connections.map((c) => [c.id, c]));
      const job = await createJobAsync(
        getCreateNewSingleTableAiGenerateJobRequest(
          {
            define: defineFormValues,
            connect: connectFormValues,
            schema: values,
          },
          account.id,
          (id) => connMap.get(id)
        )
      );
      posthog.capture('New Job Created', { jobType: 'ai-generate' });
      toast.success('Successfully created job!');

      clearNewJobSession(window.sessionStorage, sessionPrefix);

      if (job.job?.id) {
        router.push(`/${account?.name}/jobs/${job.job.id}`);
      } else {
        router.push(`/${account?.name}/jobs`);
      }
    } catch (err) {
      console.error(err);
      toast.error('Unable to create job!', {
        description: getErrorMessage(err),
      });
    }
  }

  const { data: tableConstraints, isFetching: isTableConstraintsValidating } =
    useQuery(
      getConnectionTableConstraints,
      { connectionId: connectFormValues.fkSourceConnectionId },
      { enabled: !!connectFormValues.fkSourceConnectionId }
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
      const output = await sampleRecords(
        getSampleAiGeneratedRecordsRequest({
          connect: connectFormValues,
          schema: form.getValues(),
        })
      );
      setaioutput(output.records.map((r) => fromStructToRecord(r)));
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
      form.setValue('schema', '');
      form.setValue('table', '');
      return;
    }
    if (action === 'add') {
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

          <FormField
            control={form.control}
            name="generateBatchSize"
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
