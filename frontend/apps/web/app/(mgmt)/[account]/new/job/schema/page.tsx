'use client';

import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import NosqlTable from '@/components/jobs/NosqlTable/NosqlTable';
import {
  SchemaTable,
  getAllFormErrors,
} from '@/components/jobs/SchemaTable/SchemaTable';
import { getSchemaConstraintHandler } from '@/components/jobs/SchemaTable/schema-constraint-handler';
import { setOnboardingConfig } from '@/components/onboarding-checklist/OnboardingChecklist';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { toast } from '@/components/ui/use-toast';
import { useGetAccountOnboardingConfig } from '@/libs/hooks/useGetAccountOnboardingConfig';
import { useGetConnection } from '@/libs/hooks/useGetConnection';
import { useGetConnectionSchemaMap } from '@/libs/hooks/useGetConnectionSchemaMap';
import { useGetConnectionTableConstraints } from '@/libs/hooks/useGetConnectionTableConstraints';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { validateJobMapping } from '@/libs/requests/validateJobMappings';
import { getErrorMessage } from '@/util/util';
import { SCHEMA_FORM_SCHEMA, SchemaFormValues } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  Connection,
  DatabaseColumn,
  ForeignConstraintTables,
  GetAccountOnboardingConfigResponse,
  PrimaryConstraint,
  ValidateJobMappingsResponse,
} from '@neosync/sdk';
import { useRouter } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useMemo, useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import { getOnSelectedTableToggle } from '../../../jobs/[id]/source/components/util';
import { createNewSyncJob } from '../../../jobs/util';
import JobsProgressSteps, { getJobProgressSteps } from '../JobsProgressSteps';
import { ConnectFormValues, DefineFormValues } from '../schema';

const isBrowser = () => typeof window !== 'undefined';

export interface ColumnMetadata {
  pk: { [key: string]: PrimaryConstraint };
  fk: { [key: string]: ForeignConstraintTables };
  isNullable: DatabaseColumn[];
}

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  const posthog = usePostHog();
  const { data: onboardingData, mutate } = useGetAccountOnboardingConfig(
    account?.id ?? ''
  );
  const [validateMappingsResponse, setValidateMappingsResponse] = useState<
    ValidateJobMappingsResponse | undefined
  >();

  const [isValidatingMappings, setIsValidatingMappings] = useState(false);

  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/${account?.name}/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = searchParams?.sessionId ?? '';

  // Used to complete the whole form
  const defineFormKey = `${sessionPrefix}-new-job-define`;
  const [defineFormValues] = useSessionStorage<DefineFormValues>(
    defineFormKey,
    { jobName: '' }
  );

  const connectFormKey = `${sessionPrefix}-new-job-connect`;
  const [connectFormValues] = useSessionStorage<ConnectFormValues>(
    connectFormKey,
    {
      sourceId: '',
      sourceOptions: {},
      destinations: [{ connectionId: '', destinationOptions: {} }],
    }
  );

  const schemaFormKey = `${sessionPrefix}-new-job-schema`;
  const [schemaFormData] = useSessionStorage<SchemaFormValues>(schemaFormKey, {
    mappings: [],
    connectionId: '', // hack to track if source id changes
  });

  const { data: connectionData, isLoading: isConnectionLoading } =
    useGetConnection(account?.id ?? '', connectFormValues.sourceId);

  const {
    data: connectionSchemaDataMap,
    isLoading: isSchemaMapLoading,
    isValidating: isSchemaMapValidating,
  } = useGetConnectionSchemaMap(account?.id ?? '', connectFormValues.sourceId);
  const { data: connectionsData } = useGetConnections(account?.id ?? '');
  const connections = connectionsData?.connections ?? [];

  const { data: tableConstraints, isValidating: isTableConstraintsValidating } =
    useGetConnectionTableConstraints(
      account?.id ?? '',
      connectFormValues.sourceId
    );

  const form = useForm<SchemaFormValues>({
    resolver: yupResolver<SchemaFormValues>(SCHEMA_FORM_SCHEMA),
    values: getFormValues(connectFormValues.sourceId, schemaFormData),
    context: { accountId: account?.id },
  });

  useFormPersist(`${sessionPrefix}-new-job-schema`, {
    watch: form.watch,
    setValue: form.setValue,
    storage: isBrowser() ? window.sessionStorage : undefined,
  });

  async function onSubmit(values: SchemaFormValues) {
    if (!account || !connectionData?.connection) {
      return;
    }
    if (isNosqlSource(connectionData.connection)) {
      try {
        const connMap = new Map(connections.map((c) => [c.id, c]));
        const job = await createNewSyncJob(
          {
            define: defineFormValues,
            connect: connectFormValues,
            schema: values,
            // subset: {},
          },
          account.id,
          (id) => connMap.get(id)
        );
        posthog.capture('New Job Flow Complete', {
          jobType: 'data-sync',
        });
        toast({
          title: 'Successfully created the job!',
          variant: 'success',
        });
        window.sessionStorage.removeItem(defineFormKey);
        window.sessionStorage.removeItem(connectFormKey);
        window.sessionStorage.removeItem(schemaFormKey);

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
      return;
    }
    posthog.capture('New Job Flow Schema Complete', { jobType: 'data-sync' });
    router.push(`/${account?.name}/new/job/subset?sessionId=${sessionPrefix}`);
  }

  const formMappings = form.watch('mappings');
  async function validateMappings() {
    try {
      setIsValidatingMappings(true);
      const res = await validateJobMapping(
        connectFormValues.sourceId,
        formMappings,
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

  const { append, remove, fields, update } = useFieldArray<SchemaFormValues>({
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
    const validateJobMappings = async () => {
      await validateMappings();
    };
    validateJobMappings();
  }, [selectedTables]);

  useEffect(() => {
    if (
      isSchemaMapLoading ||
      selectedTables.size > 0 ||
      !connectFormValues.sourceId
    ) {
      return;
    }
    const js = getFormValues(connectFormValues.sourceId, schemaFormData);
    setSelectedTables(
      new Set(
        js.mappings.map((mapping) => `${mapping.schema}.${mapping.table}`)
      )
    );
  }, [isSchemaMapLoading, connectFormValues.sourceId]);

  if (isConnectionLoading || isSchemaMapLoading) {
    return <SkeletonForm />;
  }
  return (
    <div className="flex flex-col gap-5">
      <OverviewContainer
        Header={
          <PageHeader
            header="Schema"
            progressSteps={
              <JobsProgressSteps
                steps={getJobProgressSteps(
                  'data-sync',
                  !isNosqlSource(connectionData?.connection ?? new Connection())
                )}
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
          {isNosqlSource(connectionData?.connection ?? new Connection({})) && (
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

          {!isNosqlSource(connectionData?.connection ?? new Connection({})) && (
            <SchemaTable
              data={formMappings}
              jobType="sync"
              constraintHandler={schemaConstraintHandler}
              schema={connectionSchemaDataMap?.schemaMap ?? {}}
              isSchemaDataReloading={isSchemaMapValidating}
              isJobMappingsValidating={isValidatingMappings}
              selectedTables={selectedTables}
              onSelectedTableToggle={onSelectedTableToggle}
              formErrors={getAllFormErrors(
                form.formState.errors,
                formMappings,
                validateMappingsResponse
              )}
              onValidate={validateMappings}
            />
          )}
          <div className="flex flex-row gap-1 justify-between">
            <Button key="back" type="button" onClick={() => router.back()}>
              Back
            </Button>
            <Button key="submit" type="submit">
              {isNosqlSource(connectionData?.connection ?? new Connection())
                ? 'Submit'
                : 'Next'}
            </Button>
          </div>
        </form>
      </Form>
    </div>
  );
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

function getFormValues(
  connectionId: string,
  existingData: SchemaFormValues | undefined
): SchemaFormValues {
  const existingMappings = existingData?.mappings ?? [];
  if (
    existingData &&
    existingMappings.length > 0 &&
    existingData.connectionId === connectionId
  ) {
    return existingData;
  }

  return {
    mappings: [],
    connectionId,
  };
}
