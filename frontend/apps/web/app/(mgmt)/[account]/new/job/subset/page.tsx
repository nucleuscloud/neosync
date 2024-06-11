'use client';

import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import SubsetOptionsForm from '@/components/jobs/Form/SubsetOptionsForm';
import EditItem from '@/components/jobs/subsets/EditItem';
import { TableRow } from '@/components/jobs/subsets/subset-table/column';
import SubsetTable from '@/components/jobs/subsets/subset-table/SubsetTable';
import {
  buildRowKey,
  buildTableRowData,
  GetColumnsForSqlAutocomplete,
} from '@/components/jobs/subsets/utils';
import { setOnboardingConfig } from '@/components/onboarding-checklist/OnboardingChecklist';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { Separator } from '@/components/ui/separator';
import { useToast } from '@/components/ui/use-toast';
import { useGetAccountOnboardingConfig } from '@/libs/hooks/useGetAccountOnboardingConfig';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { useGetConnectionTableConstraints } from '@/libs/hooks/useGetConnectionTableConstraints';
import { getErrorMessage } from '@/util/util';
import { SchemaFormValues } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  ConnectionConfig,
  GetAccountOnboardingConfigResponse,
  JobMapping,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps, { getJobProgressSteps } from '../JobsProgressSteps';
import {
  ConnectFormValues,
  DefineFormValues,
  SUBSET_FORM_SCHEMA,
  SubsetFormValues,
} from '../schema';
import { createNewJob } from './util';

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  const posthog = usePostHog();
  const { data: onboardingData, mutate } = useGetAccountOnboardingConfig(
    account?.id ?? ''
  );

  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/${account?.name}/new/job`);
    }
  }, [searchParams?.sessionId]);
  const { toast } = useToast();
  const { data: connectionsData } = useGetConnections(account?.id ?? '');
  const connections = connectionsData?.connections ?? [];

  const sessionPrefix = searchParams?.sessionId ?? '';
  const formKey = `${sessionPrefix}-new-job-subset`;

  const [subsetFormValues] = useSessionStorage<SubsetFormValues>(formKey, {
    subsets: [],
    subsetOptions: {
      subsetByForeignKeyConstraints: true,
    },
  });

  // Used to complete the whole form
  const defineFormKey = `${sessionPrefix}-new-job-define`;
  const [defineFormValues] = useSessionStorage<DefineFormValues>(
    defineFormKey,
    { jobName: '' }
  );

  // Used to complete the whole form
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
  const [schemaFormValues] = useSessionStorage<SchemaFormValues>(
    schemaFormKey,
    {
      mappings: [],
      connectionId: '',
    }
  );

  const { data: tableConstraints, isValidating: isTableConstraintsValidating } =
    useGetConnectionTableConstraints(
      account?.id ?? '',
      schemaFormValues.connectionId ?? ''
    );

  const fkConstraints = tableConstraints?.foreignKeyConstraints;
  const [rootTables, setRootTables] = useState<Set<string>>(new Set());
  useEffect(() => {
    if (!isTableConstraintsValidating && fkConstraints) {
      schemaFormValues.mappings.forEach((m) => {
        const tn = `${m.schema}.${m.table}`;
        if (!fkConstraints[tn]) {
          rootTables.add(tn);
          setRootTables(rootTables);
        }
      });
    }
  }, [fkConstraints, isTableConstraintsValidating]);

  const form = useForm({
    resolver: yupResolver<SubsetFormValues>(SUBSET_FORM_SCHEMA),
    defaultValues: subsetFormValues,
  });

  const isBrowser = () => typeof window !== 'undefined';
  useFormPersist(formKey, {
    watch: form.watch,
    setValue: form.setValue,
    storage: isBrowser() ? window.sessionStorage : undefined,
  });

  const [itemToEdit, setItemToEdit] = useState<TableRow | undefined>();

  const connectionType = connections.find(
    (item) => item.id == connectFormValues.sourceId
  );

  const dbType = getDbtype();
  connectionType?.connectionConfig?.config.case == 'mysqlConfig'
    ? 'mysql'
    : 'postgres';

  async function onSubmit(values: SubsetFormValues): Promise<void> {
    if (!account) {
      return;
    }

    try {
      const job = await createNewJob(
        {
          define: defineFormValues,
          connect: connectFormValues,
          schema: schemaFormValues,
          subset: values,
        },
        account.id,
        connections
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

  const tableRowData = buildTableRowData(
    schemaFormValues.mappings,
    rootTables,
    form.watch().subsets
  );

  function hasLocalChange(schema: string, table: string): boolean {
    const key = buildRowKey(schema, table);
    const trData = tableRowData[key];
    const svrData = subsetFormValues.subsets.find(
      (ss) => buildRowKey(ss.schema, ss.table) === key
    );
    if (!svrData && !!trData.where) {
      return true;
    }
    return trData.where !== svrData?.whereClause;
  }

  function onLocalRowReset(schema: string, table: string): void {
    const key = buildRowKey(schema, table);
    const idx = form
      .getValues()
      .subsets.findIndex(
        (item) => buildRowKey(item.schema, item.table) === key
      );
    if (idx >= 0) {
      const svrData = subsetFormValues.subsets.find(
        (ss) => buildRowKey(ss.schema, ss.table) === key
      );

      form.setValue(`subsets.${idx}`, {
        schema: schema,
        table: table,
        whereClause: svrData?.whereClause ?? undefined,
      });
    }
  }

  return (
    <div className="px-12 md:px-24 lg:px-32 flex flex-col gap-5">
      <OverviewContainer
        Header={
          <PageHeader
            header="Subset"
            progressSteps={
              <JobsProgressSteps
                steps={getJobProgressSteps('data-sync', true)}
                stepName={'subset'}
              />
            }
          />
        }
        containerClassName="connect-page"
      >
        <div />
      </OverviewContainer>
      {dbType === 'invalid' && (
        <Alert variant="warning">
          <ExclamationTriangleIcon className="h-4 w-4" />
          <AlertTitle>Heads up!</AlertTitle>
          <AlertDescription>
            Subsetting is not currently enabled for NoSQL jobs. You may proceed
            with the creation of this job while we continue to work on NoSQL
            subsetting.
          </AlertDescription>
        </Alert>
      )}

      {dbType !== 'invalid' && (
        <div className="flex flex-col gap-4">
          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(onSubmit)}
              className="flex flex-col gap-8"
            >
              <div>
                <SubsetOptionsForm maxColNum={2} />
              </div>
              <div className="flex flex-col gap-2">
                <div>
                  <SubsetTable
                    data={Object.values(tableRowData)}
                    onEdit={(schema, table) => {
                      const key = buildRowKey(schema, table);
                      if (tableRowData[key]) {
                        // make copy so as to not edit in place
                        setItemToEdit({
                          ...tableRowData[key],
                        });
                      }
                    }}
                    hasLocalChange={hasLocalChange}
                    onReset={onLocalRowReset}
                  />
                </div>
                <div className="my-4">
                  <Separator />
                </div>
                <div>
                  <EditItem
                    connectionId={connectFormValues.sourceId}
                    item={itemToEdit}
                    onItem={setItemToEdit}
                    onCancel={() => setItemToEdit(undefined)}
                    columns={GetColumnsForSqlAutocomplete(
                      schemaFormValues?.mappings.map((row) => {
                        return new JobMapping({
                          schema: row.schema,
                          table: row.table,
                          column: row.column,
                        });
                      }),
                      itemToEdit
                    )}
                    onSave={() => {
                      if (!itemToEdit) {
                        return;
                      }
                      const key = buildRowKey(
                        itemToEdit.schema,
                        itemToEdit.table
                      );
                      const idx = form
                        .getValues()
                        .subsets.findIndex(
                          (item) => buildRowKey(item.schema, item.table) === key
                        );
                      if (idx >= 0) {
                        form.setValue(`subsets.${idx}`, {
                          schema: itemToEdit.schema,
                          table: itemToEdit.table,
                          whereClause: itemToEdit.where,
                        });
                      } else {
                        form.setValue(
                          `subsets`,
                          form.getValues().subsets.concat({
                            schema: itemToEdit.schema,
                            table: itemToEdit.table,
                            whereClause: itemToEdit.where,
                          })
                        );
                      }
                      setItemToEdit(undefined);
                    }}
                    dbType={dbType}
                  />
                </div>

                <div className="my-6">
                  <Separator />
                </div>
                <div className="flex flex-row gap-1 justify-between">
                  <Button
                    key="back"
                    type="button"
                    onClick={() => router.back()}
                  >
                    Back
                  </Button>
                  <Button key="submit" type="submit">
                    Save
                  </Button>
                </div>
              </div>
            </form>
          </Form>
        </div>
      )}
    </div>
  );
}

function getDbtype(
  options?: ConnectionConfig
): 'mysql' | 'postgres' | 'invalid' {
  switch (options?.config.case) {
    case 'pgConfig':
      return 'postgres';
    case 'mysqlConfig':
      return 'mysql';
    default:
      return 'invalid';
  }
}
