'use client';

import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import NosqlTable from '@/components/jobs/NosqlTable/NosqlTable';
import {
  getAllFormErrors,
  SchemaTable,
} from '@/components/jobs/SchemaTable/SchemaTable';
import { getSchemaConstraintHandler } from '@/components/jobs/SchemaTable/schema-constraint-handler';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { toast } from '@/components/ui/use-toast';
import { useGetConnectionSchemaMap } from '@/libs/hooks/useGetConnectionSchemaMap';
import { validateJobMapping } from '@/libs/requests/validateJobMappings';
import { getSingleOrUndefined } from '@/libs/utils';
import { getErrorMessage } from '@/util/util';
import {
  SchemaFormValues,
  VirtualForeignConstraintFormValues,
} from '@/yup-validations/jobs';
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  Connection,
  DatabaseColumn,
  ForeignConstraintTables,
  GetAccountOnboardingConfigResponse,
  PrimaryConstraint,
  ValidateJobMappingsResponse,
  VirtualForeignConstraint,
  VirtualForeignKey,
} from '@neosync/sdk';
import {
  createJob,
  getAccountOnboardingConfig,
  getConnection,
  getConnections,
  getConnectionTableConstraints,
  setAccountOnboardingConfig,
} from '@neosync/sdk/connectquery';
import { useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useMemo, useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import { getOnSelectedTableToggle } from '../../../jobs/[id]/source/components/util';
import {
  clearNewJobSession,
  getCreateNewSyncJobRequest,
  getNewJobSessionKeys,
} from '../../../jobs/util';
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
  const { data: onboardingData } = useQuery(
    getAccountOnboardingConfig,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const queryclient = useQueryClient();
  const { mutateAsync: setOnboardingConfig } = useMutation(
    setAccountOnboardingConfig
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

  const sessionPrefix = getSingleOrUndefined(searchParams?.sessionId) ?? '';
  const sessionKeys = getNewJobSessionKeys(sessionPrefix);

  // Used to complete the whole form
  const defineFormKey = sessionKeys.global.define;
  const [defineFormValues] = useSessionStorage<DefineFormValues>(
    defineFormKey,
    { jobName: '' }
  );

  const connectFormKey = sessionKeys.dataSync.connect;
  const [connectFormValues] = useSessionStorage<ConnectFormValues>(
    connectFormKey,
    {
      sourceId: '',
      sourceOptions: {},
      destinations: [{ connectionId: '', destinationOptions: {} }],
    }
  );

  const schemaFormKey = sessionKeys.dataSync.schema;
  const [schemaFormData] = useSessionStorage<SchemaFormValues>(schemaFormKey, {
    mappings: [],
    connectionId: '', // hack to track if source id changes
  });

  const { data: connectionData, isLoading: isConnectionLoading } = useQuery(
    getConnection,
    { id: connectFormValues.sourceId },
    { enabled: !!connectFormValues.sourceId }
  );

  const {
    data: connectionSchemaDataMap,
    isLoading: isSchemaMapLoading,
    isValidating: isSchemaMapValidating,
  } = useGetConnectionSchemaMap(account?.id ?? '', connectFormValues.sourceId);
  const { data: connectionsData } = useQuery(
    getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const connections = connectionsData?.connections ?? [];

  const { data: tableConstraints, isFetching: isTableConstraintsValidating } =
    useQuery(
      getConnectionTableConstraints,
      { connectionId: connectFormValues.sourceId },
      { enabled: !!connectFormValues.sourceId }
    );

  const { mutateAsync: createNewSyncJob } = useMutation(createJob);

  const form = useForm<SchemaFormValues>({
    resolver: yupResolver<SchemaFormValues>(SchemaFormValues),
    values: getFormValues(connectFormValues.sourceId, schemaFormData),
    context: { accountId: account?.id },
  });

  useFormPersist(schemaFormKey, {
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
          getCreateNewSyncJobRequest(
            {
              define: defineFormValues,
              connect: connectFormValues,
              schema: values,
              // subset: {},
            },
            account.id,
            (id) => connMap.get(id)
          )
        );
        posthog.capture('New Job Flow Complete', {
          jobType: 'data-sync',
        });
        toast({
          title: 'Successfully created the job!',
          variant: 'success',
        });

        clearNewJobSession(window.sessionStorage, sessionPrefix);

        // updates the onboarding data
        if (!onboardingData?.config?.hasCreatedJob) {
          try {
            const resp = await setOnboardingConfig({
              accountId: account.id,
              config: {
                hasCreatedSourceConnection:
                  onboardingData?.config?.hasCreatedSourceConnection ?? true,
                hasCreatedDestinationConnection:
                  onboardingData?.config?.hasCreatedDestinationConnection ??
                  true,
                hasCreatedJob: true,
                hasInvitedMembers:
                  onboardingData?.config?.hasInvitedMembers ?? true,
              },
            });
            queryclient.setQueryData(
              createConnectQueryKey(getAccountOnboardingConfig, {
                accountId: account.id,
              }),
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
  const formVirtualForeignKeys = form.watch('virtualForeignKeys');
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

  async function validateVirtualForeignKeys(
    vfks: VirtualForeignConstraintFormValues[]
  ) {
    try {
      setIsValidatingMappings(true);
      const res = await validateJobMapping(
        connectFormValues.sourceId,
        formMappings,
        account?.id || '',
        vfks
      );
      setValidateMappingsResponse(res);
    } catch (error) {
      console.error('Failed to validate virtual foreign keys:', error);
      toast({
        title: 'Unable to validate virtual foreign keys',
        variant: 'destructive',
      });
    } finally {
      setIsValidatingMappings(false);
    }
  }

  const schemaConstraintHandler = useMemo(() => {
    const virtualForeignKeys = formVirtualForeignKeys?.map((v) => {
      return new VirtualForeignConstraint({
        schema: v.schema,
        table: v.table,
        columns: v.columns,
        foreignKey: new VirtualForeignKey({
          schema: v.foreignKey.schema,
          table: v.foreignKey.table,
          columns: v.foreignKey.columns,
        }),
      });
    });
    return getSchemaConstraintHandler(
      connectionSchemaDataMap?.schemaMap ?? {},
      tableConstraints?.primaryKeyConstraints ?? {},
      tableConstraints?.foreignKeyConstraints ?? {},
      tableConstraints?.uniqueConstraints ?? {},
      virtualForeignKeys ?? []
    );
  }, [
    isSchemaMapValidating,
    isTableConstraintsValidating,
    formVirtualForeignKeys,
  ]);
  const [selectedTables, setSelectedTables] = useState<Set<string>>(new Set());

  const { append, remove, fields, update } = useFieldArray<SchemaFormValues>({
    control: form.control,
    name: 'mappings',
  });

  const { append: appendVfk, remove: removeVfk } =
    useFieldArray<SchemaFormValues>({
      control: form.control,
      name: 'virtualForeignKeys',
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
    if (!connectFormValues.sourceId || !account?.id) {
      return;
    }
    const validateJobMappings = async () => {
      await validateMappings();
    };
    validateJobMappings();
  }, [selectedTables, connectFormValues.sourceId, account?.id]);

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

  async function addVirtualForeignKey(vfk: VirtualForeignConstraintFormValues) {
    appendVfk(vfk);
    const vfks = [vfk, ...(formVirtualForeignKeys || [])];
    await validateVirtualForeignKeys(vfks);
  }

  async function removeVirtualForeignKey(index: number) {
    const newVfks: VirtualForeignConstraintFormValues[] = [];
    formVirtualForeignKeys?.forEach((vfk, idx) => {
      if (idx != index) {
        newVfks.push(vfk);
      }
    });
    removeVfk(index);
    await validateVirtualForeignKeys(newVfks);
  }

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
              virtualForeignKeys={formVirtualForeignKeys}
              addVirtualForeignKey={addVirtualForeignKey}
              removeVirtualForeignKey={removeVirtualForeignKey}
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
    virtualForeignKeys: [],
    connectionId,
  };
}
