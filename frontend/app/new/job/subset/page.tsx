'use client';

import EditItem from '@/components/jobs/subsets/EditItem';
import SubsetTable from '@/components/jobs/subsets/subset-table/SubsetTable';
import { TableRow } from '@/components/jobs/subsets/subset-table/column';
import {
  buildRowKey,
  buildTableRowData,
} from '@/components/jobs/subsets/utils';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { Separator } from '@/components/ui/separator';
import { useToast } from '@/components/ui/use-toast';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
<<<<<<< HEAD
=======
import { useGetSystemTransformers } from '@/libs/hooks/useGetSystemTransformers';
import { useGetUserDefinedTransformers } from '@/libs/hooks/useGetUserDefinedTransformers';
>>>>>>> f11365ea (working job)
import { Connection } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import {
  CreateJobRequest,
  CreateJobResponse,
  JobDestination,
  JobMapping,
  JobMappingTransformer,
  JobSource,
  JobSourceOptions,
  MysqlSourceConnectionOptions,
  PostgresSourceConnectionOptions,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import {
  SchemaFormValues,
  toJobDestinationOptions,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps, { DATA_SYNC_STEPS } from '../JobsProgressSteps';
import {
  ConnectFormValues,
  DefineFormValues,
  FormValues,
  SUBSET_FORM_SCHEMA,
  SubsetFormValues,
} from '../schema';

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/new/job`);
    }
  }, [searchParams?.sessionId]);
  const { toast } = useToast();
  const { data: connectionsData } = useGetConnections(account?.id ?? '');
  const connections = connectionsData?.connections ?? [];

  const sessionPrefix = searchParams?.sessionId ?? '';
  const formKey = `${sessionPrefix}-new-job-subset`;

  const [subsetFormValues] = useSessionStorage<SubsetFormValues>(formKey, {
    subsets: [],
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
    }
  );

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
      window.sessionStorage.removeItem(defineFormKey);
      window.sessionStorage.removeItem(connectFormKey);
      window.sessionStorage.removeItem(schemaFormKey);
      window.sessionStorage.removeItem(formKey);
      if (job.job?.id) {
        router.push(`/jobs/${job.job.id}`);
      } else {
        router.push(`/jobs`);
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
    <div className="px-12 md:px-24 lg:px-32 flex flex-col gap-20">
      <div className="mt-10">
        <JobsProgressSteps steps={DATA_SYNC_STEPS} stepName={'subset'} />
      </div>
      <div className="flex flex-col gap-4">
        <div>
          <h2 className="text-1xl font-bold tracking-tight">
            Set table subset rules by pressing the edit button and filling out
            the form below
          </h2>
        </div>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="flex flex-col gap-2"
          >
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
                onSave={() => {
                  if (!itemToEdit) {
                    return;
                  }
                  const key = buildRowKey(itemToEdit.schema, itemToEdit.table);
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
              />
            </div>
            <div className="my-6">
              <Separator />
            </div>
            <div className="flex flex-row gap-1 justify-between">
              <Button key="back" type="button" onClick={() => router.back()}>
                Back
              </Button>
              <Button key="submit" type="submit">
                Save
              </Button>
            </div>
          </form>
        </Form>
      </div>
    </div>
  );
}

async function createNewJob(
  formData: FormValues,
  accountId: string,
  connections: Connection[]
): Promise<CreateJobResponse> {
  const connectionIdMap = new Map(
    connections.map((connection) => [connection.id, connection])
  );
  const sourceConnection = connections.find(
    (c) => c.id == formData.connect.sourceId
  );

  console.log('formData', formData.schema.mappings);

  const body = new CreateJobRequest({
    accountId,
    jobName: formData.define.jobName,
    cronSchedule: formData.define.cronSchedule,
    initiateJobRun: formData.define.initiateJobRun,
    mappings: formData.schema.mappings.map((m) => {
      return new JobMapping({
        schema: m.schema,
        table: m.table,
        column: m.column,
        transformer: JobMappingTransformer.fromJson(m.transformer),
      });
    }),
    source: new JobSource({
      options: toJobSourceOptions(formData, sourceConnection),
    }),
    destinations: formData.connect.destinations.map((d) => {
      return new JobDestination({
        connectionId: d.connectionId,
        options: toJobDestinationOptions(
          d,
          connectionIdMap.get(d.connectionId)
        ),
      });
    }),
  });

  function toJobSourceOptions(
    values: FormValues,
    connection?: Connection
  ): JobSourceOptions {
    if (!connection) {
      return new JobSourceOptions();
    }
    switch (connection.connectionConfig?.config.case) {
      case 'pgConfig':
        return new JobSourceOptions({
          config: {
            case: 'postgres',
            value: new PostgresSourceConnectionOptions({
              connectionId: formData.connect.sourceId,
              haltOnNewColumnAddition:
                values.connect.sourceOptions.haltOnNewColumnAddition,
            }),
          },
        });
      case 'mysqlConfig':
        return new JobSourceOptions({
          config: {
            case: 'mysql',
            value: new MysqlSourceConnectionOptions({
              connectionId: formData.connect.sourceId,
              haltOnNewColumnAddition:
                values.connect.sourceOptions.haltOnNewColumnAddition,
            }),
          },
        });
      default:
        throw new Error('unsupported connection type');
    }
  }

  const res = await fetch(`/api/jobs`, {
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
