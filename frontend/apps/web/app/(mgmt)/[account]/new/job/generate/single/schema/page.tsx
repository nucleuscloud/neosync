'use client';

import FormPersist from '@/app/(mgmt)/FormPersist';
import { useOnApplyDefaultClick } from '@/app/(mgmt)/[account]/jobs/[id]/source/components/useOnApplyDefaultClick';
import { useOnImportMappings } from '@/app/(mgmt)/[account]/jobs/[id]/source/components/useOnImportMappings';
import { useOnTransformerBulkUpdateClick } from '@/app/(mgmt)/[account]/jobs/[id]/source/components/useOnTransformerBulkUpdateClick';
import {
  getFilteredTransformersForBulkSet,
  getOnSelectedTableToggle,
} from '@/app/(mgmt)/[account]/jobs/[id]/source/components/util';
import {
  clearNewJobSession,
  getCreateNewSingleTableGenerateJobRequest,
  getNewJobSessionKeys,
  toSingleTableGenerateJobSource,
  validateJobMapping,
} from '@/app/(mgmt)/[account]/jobs/util';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import {
  getAllFormErrors,
  SchemaTable,
} from '@/components/jobs/SchemaTable/SchemaTable';
import { getSchemaConstraintHandler } from '@/components/jobs/SchemaTable/schema-constraint-handler';
import { TransformerResult } from '@/components/jobs/SchemaTable/transformer-handler';
import { getTransformerFilter } from '@/components/jobs/SchemaTable/util';
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
import { useGetTransformersHandler } from '@/libs/hooks/useGetTransformersHandler';
import { getSingleOrUndefined } from '@/libs/utils';
import { getErrorMessage, getTransformerFromField } from '@/util/util';
import { JobMappingTransformerForm } from '@/yup-validations/jobs';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  ConnectionDataService,
  ConnectionService,
  JobService,
  ValidateJobMappingsResponse,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, use, useEffect, useMemo, useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps, {
  getJobProgressSteps,
} from '../../../JobsProgressSteps';
import {
  DefineFormValues,
  SingleTableConnectFormValues,
  SingleTableSchemaFormValues,
} from '../../../job-form-validations';

export default function Page(props: PageProps): ReactElement {
  const searchParams = use(props.searchParams);
  const { account } = useAccount();
  const router = useRouter();
  const posthog = usePostHog();

  const [validateMappingsResponse, setValidateMappingsResponse] = useState<
    ValidateJobMappingsResponse | undefined
  >();

  const [isValidatingMappings, setIsValidatingMappings] = useState(false);

  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/${account?.name}/new/job`);
    }
  }, [searchParams?.sessionId]);
  const { data: connectionsData } = useQuery(
    ConnectionService.method.getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const connections = connectionsData?.connections ?? [];

  const sessionPrefix = getSingleOrUndefined(searchParams?.sessionId) ?? '';
  const sessionKeys = getNewJobSessionKeys(sessionPrefix);
  // Used to complete the whole form
  const defineFormKey = sessionKeys.global.define;
  const [defineFormValues] = useSessionStorage<DefineFormValues>(
    defineFormKey,
    { jobName: '' }
  );
  const connectFormKey = sessionKeys.generate.connect;
  const [connectFormValues] = useSessionStorage<SingleTableConnectFormValues>(
    connectFormKey,
    {
      fkSourceConnectionId: '',
      destination: {
        connectionId: '',
        destinationOptions: {},
      },
    }
  );

  const formKey = sessionKeys.generate.schema;
  const [schemaFormData] = useSessionStorage<SingleTableSchemaFormValues>(
    formKey,
    {
      mappings: [],
      numRows: 10,
    }
  );

  const {
    data: connectionSchemaDataMap,
    isLoading: isSchemaMapLoading,
    isFetching: isSchemaMapValidating,
  } = useQuery(
    ConnectionDataService.method.getConnectionSchemaMap,
    { connectionId: connectFormValues.fkSourceConnectionId },
    { enabled: !!connectFormValues.fkSourceConnectionId }
  );

  const { mutateAsync: createJobAsync } = useMutation(
    JobService.method.createJob
  );

  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver<SingleTableSchemaFormValues>(
      SingleTableSchemaFormValues
    ),
    values: schemaFormData,
    context: { accountId: account?.id },
  });

  const [isClient, setIsClient] = useState(false);
  useEffect(() => {
    setIsClient(true);
  }, []);

  const { mutateAsync: validateJobMappingsAsync } = useMutation(
    JobService.method.validateJobMappings
  );

  async function onSubmit(values: SingleTableSchemaFormValues) {
    if (!account) {
      return;
    }
    try {
      const connMap = new Map(connections.map((c) => [c.id, c]));
      const job = await createJobAsync(
        getCreateNewSingleTableGenerateJobRequest(
          {
            define: defineFormValues,
            connect: connectFormValues,
            schema: values,
          },
          account.id,
          (id) => connMap.get(id)
        )
      );
      posthog.capture('New Job Created', {
        jobType: 'generate',
      });

      toast.success('Successfully created job!');

      clearNewJobSession(window.sessionStorage, sessionPrefix);
      if (job.job?.id) {
        router.push(`/${account?.name}/jobs/${job.job.id}`);
      } else {
        router.push(`/${account?.name}/jobs`);
      }
    } catch (err) {
      console.error(err);
      toast.error('Unable to create job', {
        description: getErrorMessage(err),
      });
    }
  }
  const formMappings = form.watch('mappings');

  async function validateMappings() {
    if (formMappings.length === 0) {
      return;
    }
    try {
      setIsValidatingMappings(true);
      const res = await validateJobMapping(
        formMappings,
        account?.id || '',
        [],
        validateJobMappingsAsync,
        toSingleTableGenerateJobSource({
          connect: connectFormValues,
          schema: form.getValues(),
        })
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

  const { data: tableConstraints, isFetching: isTableConstraintsValidating } =
    useQuery(
      ConnectionDataService.method.getConnectionTableConstraints,
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
  const { append, remove, fields, update } =
    useFieldArray<SingleTableSchemaFormValues>({
      control: form.control,
      name: 'mappings',
    });

  const onSelectedTableToggle = getOnSelectedTableToggle(
    connectionSchemaDataMap?.schemaMap ?? {},
    selectedTables,
    setSelectedTables,
    fields,
    remove,
    append
  );

  useEffect(() => {
    if (!connectFormValues.fkSourceConnectionId || !account?.id) {
      return;
    }
    const validateJobMappings = async () => {
      await validateMappings();
    };
    validateJobMappings();
  }, [selectedTables, connectFormValues.fkSourceConnectionId, account?.id]);

  useEffect(() => {
    if (isSchemaMapLoading || selectedTables.size > 0) {
      return;
    }
    const js = schemaFormData;
    setSelectedTables(
      new Set(
        js.mappings.map((mapping) => `${mapping.schema}.${mapping.table}`)
      )
    );
  }, [isSchemaMapLoading]);

  const { handler } = useGetTransformersHandler(account?.id ?? '');

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
    const row = formMappings[idx];
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

  return (
    <div className="flex flex-col gap-5">
      <FormPersist formKey={formKey} form={form} />
      <OverviewContainer
        Header={
          <PageHeader
            header="Schema"
            progressSteps={
              <JobsProgressSteps
                steps={getJobProgressSteps('generate-table')}
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
        <form onSubmit={form.handleSubmit(onSubmit)}>
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

          {isClient && (
            <SchemaTable
              data={formMappings}
              constraintHandler={schemaConstraintHandler}
              schema={connectionSchemaDataMap?.schemaMap ?? {}}
              isSchemaDataReloading={isSchemaMapValidating}
              jobType={'generate'}
              selectedTables={selectedTables}
              onSelectedTableToggle={onSelectedTableToggle}
              formErrors={getAllFormErrors(
                form.formState.errors,
                formMappings,
                validateMappingsResponse
              )}
              onValidate={validateMappings}
              isJobMappingsValidating={isValidatingMappings}
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
              hasMissingSourceColumnMappings={false}
              onRemoveMissingSourceColumnMappings={() => {}}
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
          <div className="flex flex-row gap-1 justify-between pt-10">
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
