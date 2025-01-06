'use client';

import FormPersist from '@/app/(mgmt)/FormPersist';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import SubsetOptionsForm from '@/components/jobs/Form/SubsetOptionsForm';
import EditItem from '@/components/jobs/subsets/EditItem';
import EditItemDialog from '@/components/jobs/subsets/EditItemDialog';
import {
  SUBSET_TABLE_COLUMNS,
  SubsetTableRow,
} from '@/components/jobs/subsets/SubsetTable/Columns';
import SubsetTable from '@/components/jobs/subsets/SubsetTable/SubsetTable';
import {
  buildRowKey,
  buildTableRowData,
  getColumnsForSqlAutocomplete,
  isValidSubsetType,
} from '@/components/jobs/subsets/utils';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { getSingleOrUndefined } from '@/libs/utils';
import { getErrorMessage } from '@/util/util';
import { SchemaFormValues } from '@/yup-validations/jobs';
import { create } from '@bufbuild/protobuf';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  ConnectionConfigSchema,
  ConnectionDataService,
  ConnectionService,
  JobService,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useMemo, useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { useSessionStorage } from 'usehooks-ts';
import { getConnectionType } from '../../../connections/util';
import { showSubsetOptions } from '../../../jobs/[id]/subsets/components/SubsetCard';
import {
  clearNewJobSession,
  getCreateNewSyncJobRequest,
  getNewJobSessionKeys,
} from '../../../jobs/util';
import {
  ConnectFormValues,
  DefineFormValues,
  SubsetFormValues,
} from '../job-form-validations';
import JobsProgressSteps, { getJobProgressSteps } from '../JobsProgressSteps';

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  const posthog = usePostHog();

  const [isDialogOpen, setIsDialogOpen] = useState(false);

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

  const formKey = sessionKeys.dataSync.subset;
  const [subsetFormValues] = useSessionStorage<SubsetFormValues>(formKey, {
    subsets: [],
    subsetOptions: {
      subsetByForeignKeyConstraints: true,
    },
  });

  // Used to complete the whole form
  const defineFormKey = sessionKeys.global.define;
  const [defineFormValues] = useSessionStorage<DefineFormValues>(
    defineFormKey,
    { jobName: '' }
  );

  // Used to complete the whole form
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
  const [schemaFormValues] = useSessionStorage<SchemaFormValues>(
    schemaFormKey,
    {
      mappings: [],
      connectionId: '',
      destinationOptions: [],
    }
  );

  const { data: tableConstraints, isFetching: isTableConstraintsValidating } =
    useQuery(
      ConnectionDataService.method.getConnectionTableConstraints,
      { connectionId: schemaFormValues.connectionId },
      { enabled: !!schemaFormValues.connectionId }
    );

  const { mutateAsync: createNewSyncJob } = useMutation(
    JobService.method.createJob
  );

  const fkConstraints = tableConstraints?.foreignKeyConstraints;
  const [rootTables, setRootTables] = useState<Set<string>>(new Set());
  useEffect(() => {
    if (!isTableConstraintsValidating && fkConstraints) {
      const newRootTables = new Set(rootTables);
      schemaFormValues.mappings.forEach((m) => {
        const tn = `${m.schema}.${m.table}`;
        if (!fkConstraints[tn]) {
          newRootTables.add(tn);
        }
      });
      setRootTables(newRootTables);
    }
  }, [fkConstraints, isTableConstraintsValidating]);

  const form = useForm({
    resolver: yupResolver<SubsetFormValues>(SubsetFormValues),
    defaultValues: subsetFormValues,
  });

  const [itemToEdit, setItemToEdit] = useState<SubsetTableRow | undefined>();

  const connection = connections.find(
    (item) => item.id == connectFormValues.sourceId
  );

  const connectionType = getConnectionType(
    connection?.connectionConfig ?? create(ConnectionConfigSchema, {})
  );

  async function onSubmit(values: SubsetFormValues): Promise<void> {
    if (!account) {
      return;
    }

    try {
      const connMap = new Map(connections.map((c) => [c.id, c]));
      const job = await createNewSyncJob(
        getCreateNewSyncJobRequest(
          {
            define: defineFormValues,
            connect: connectFormValues,
            schema: schemaFormValues,
            subset: values,
          },
          account.id,
          (id) => connMap.get(id)
        )
      );
      posthog.capture('New Job Flow Complete', {
        jobType: 'data-sync',
      });
      toast.success('Successfully created the job!');
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

  const formSubsets = form.watch().subsets; // ensures that all form changes cause a re-render since stuff happens outside of the form that depends on the form values
  const { update: updateSubsetsFormValues, append: addSubsetsFormValues } =
    useFieldArray({
      control: form.control,
      name: 'subsets',
    });

  const tableRowData = useMemo(() => {
    return buildTableRowData(
      schemaFormValues.mappings,
      rootTables,
      formSubsets
    );
  }, [schemaFormValues.mappings, rootTables, formSubsets]);

  const sqlAutocompleteColumns = useMemo(() => {
    console.log('re-rendering auto complete columns');
    return getColumnsForSqlAutocomplete(
      schemaFormValues?.mappings ?? [],
      itemToEdit?.schema ?? '',
      itemToEdit?.table ?? ''
    );
  }, [schemaFormValues?.mappings, itemToEdit?.schema, itemToEdit?.table]);

  function hasLocalChange(
    _rowIdx: number,
    schema: string,
    table: string
  ): boolean {
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

  function onLocalRowReset(
    _rowIdx: number,
    schema: string,
    table: string
  ): void {
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
      updateSubsetsFormValues(idx, {
        schema: schema,
        table: table,
        whereClause: svrData?.whereClause ?? undefined,
      });
      form.trigger();
    }
  }
  console.log('re-rendering subset table');
  return (
    <div className="px-12 md:px-24 lg:px-32 flex flex-col gap-5">
      <FormPersist formKey={formKey} form={form} />
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
      {!isValidSubsetType(connectionType) && (
        <Alert variant="warning">
          <ExclamationTriangleIcon className="h-4 w-4" />
          <AlertTitle>Heads up!</AlertTitle>
          <AlertDescription>
            Subsetting is not currently enabled for this connection type. You
            may proceed with the creation of this job while we continue to work
            on subsetting for this connection.
          </AlertDescription>
        </Alert>
      )}

      {isValidSubsetType(connectionType) && (
        <div className="flex flex-col gap-4">
          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(onSubmit)}
              className="flex flex-col gap-2"
            >
              <div>
                {showSubsetOptions(connectionType) && (
                  <SubsetOptionsForm maxColNum={2} />
                )}
              </div>
              <div className="flex flex-col">
                <div>
                  <SubsetTable
                    data={Object.values(tableRowData)}
                    columns={SUBSET_TABLE_COLUMNS}
                    onEdit={(_rowIdx, schema, table) => {
                      setIsDialogOpen(true);
                      const key = buildRowKey(schema, table);
                      if (tableRowData[key]) {
                        setItemToEdit({
                          ...tableRowData[key],
                        });
                      }
                    }}
                    hasLocalChange={hasLocalChange}
                    onReset={onLocalRowReset}
                  />
                </div>
                <EditItemDialog
                  open={isDialogOpen}
                  onOpenChange={setIsDialogOpen}
                  body={
                    <EditItem
                      connectionId={connectFormValues.sourceId}
                      item={itemToEdit}
                      onItem={setItemToEdit}
                      onCancel={() => {
                        setItemToEdit(undefined);
                        setIsDialogOpen(false);
                      }}
                      columns={sqlAutocompleteColumns}
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
                            (item) =>
                              buildRowKey(item.schema, item.table) === key
                          );
                        if (idx >= 0) {
                          updateSubsetsFormValues(idx, {
                            schema: itemToEdit.schema,
                            table: itemToEdit.table,
                            whereClause: itemToEdit.where,
                          });
                        } else {
                          addSubsetsFormValues({
                            schema: itemToEdit.schema,
                            table: itemToEdit.table,
                            whereClause: itemToEdit.where,
                          });
                        }
                        setItemToEdit(undefined);
                        setIsDialogOpen(false);
                      }}
                      connectionType={connectionType}
                    />
                  }
                />
                <div className="flex flex-row gap-1 justify-between pt-10">
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
