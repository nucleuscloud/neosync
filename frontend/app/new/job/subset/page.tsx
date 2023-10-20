'use client';

import ButtonText from '@/components/ButtonText';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { Separator } from '@/components/ui/separator';
import { Textarea } from '@/components/ui/textarea';
import { useToast } from '@/components/ui/use-toast';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { Connection } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import {
  CreateJobRequest,
  CreateJobResponse,
  JobDestination,
  JobMapping,
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
import { toTransformerConfigOptions } from '@/yup-validations/transformers';
import { yupResolver } from '@hookform/resolvers/yup';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import {
  DefineFormValues,
  FlowFormValues,
  FormValues,
  SUBSET_FORM_SCHEMA,
  SubsetFormValues,
} from '../schema';
import { TableRow, getColumns } from './subset-table/column';
import { DataTable } from './subset-table/data-table';

export default function Page({ searchParams }: PageProps): ReactElement {
  const account = useAccount();
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
  const sessionKey = `${sessionPrefix}-new-job-subset`;

  const [subsetFormValues] = useSessionStorage<SubsetFormValues>(sessionKey, {
    subsets: [],
  });

  // Used to complete the whole form
  const [defineFormValues] = useSessionStorage<DefineFormValues>(
    `${sessionPrefix}-new-job-define`,
    { jobName: '' }
  );

  // Used to complete the whole form
  const [flowFormValues] = useSessionStorage<FlowFormValues>(
    `${sessionPrefix}-new-job-flow`,
    {
      sourceId: '',
      sourceOptions: {},
      destinations: [{ connectionId: '', destinationOptions: {} }],
    }
  );

  const [schemaFormValues] = useSessionStorage<SchemaFormValues>(
    `${sessionPrefix}-new-job-schema`,
    {
      mappings: [],
    }
  );

  const form = useForm({
    resolver: yupResolver<SubsetFormValues>(SUBSET_FORM_SCHEMA),
    defaultValues: subsetFormValues,
  });

  const isBrowser = () => typeof window !== 'undefined';
  useFormPersist(sessionKey, {
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
          flow: flowFormValues,
          schema: schemaFormValues,
          subset: values,
        },
        account.id,
        connections
      );
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
    form.getValues().subsets
  );

  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Create a new Job"
          description="Further subset your source connection tables to reduce the amount of data translated to your destination(s)"
        />
      }
    >
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
              />
            </div>
            <div className="my-4">
              <Separator />
            </div>
            <div>
              <EditItem
                item={itemToEdit}
                onItem={setItemToEdit}
                onCancel={() => setItemToEdit(undefined)}
                onSave={() => {
                  if (!itemToEdit) {
                    return;
                  }
                  const idx = form
                    .getValues()
                    .subsets.findIndex(
                      (item) =>
                        buildRowKey(item.schema, item.table) ===
                        buildRowKey(
                          itemToEdit?.schema ?? '',
                          itemToEdit?.table ?? ''
                        )
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
    </OverviewContainer>
  );
}

interface SubsetTableProps {
  data: TableRow[];
  onEdit(schema: string, table: string): void;
}

function SubsetTable(props: SubsetTableProps): ReactElement {
  const { data, onEdit } = props;

  const columns = getColumns({ onEdit });

  return <DataTable columns={columns} data={data} />;
}

function buildTableRowData(
  mappings: SchemaFormValues['mappings'],
  existingSubsets: SubsetFormValues['subsets']
): Record<string, TableRow> {
  const tableMap: Record<string, TableRow> = {};

  mappings.forEach((mapping) => {
    const key = buildRowKey(mapping.schema, mapping.table);
    tableMap[key] = { schema: mapping.schema, table: mapping.table };
  });
  existingSubsets.forEach((subset) => {
    const key = buildRowKey(subset.schema, subset.table);
    if (tableMap[key]) {
      tableMap[key].where = subset.whereClause;
    }
  });

  return tableMap;
}

function buildRowKey(schema: string, table: string): string {
  return `${schema}.${table}`;
}

interface EditItemProps {
  item?: TableRow;
  onItem(item?: TableRow): void;
  onSave(): void;
  onCancel(): void;
}
function EditItem(props: EditItemProps): ReactElement {
  const { item, onItem, onSave, onCancel } = props;

  function onWhereChange(value: string): void {
    if (!item) {
      return;
    }
    onItem({ ...item, where: value });
  }
  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-row justify-between">
        <div className="flex flex-row gap-4">
          <div className="flex flex-row gap-2 items-center">
            <span className="font-semibold tracking-tight">Schema</span>
            <Badge
              className="px-4 py-2"
              variant={item?.schema ? 'outline' : 'secondary'}
            >
              {item?.schema ?? ''}
            </Badge>
          </div>
          <div className="flex flex-row gap-2 items-center">
            <span className="font-semibold tracking-tight">Table</span>
            <Badge
              className="px-4 py-2"
              variant={item?.table ? 'outline' : 'secondary'}
            >
              {item?.table ?? ''}
            </Badge>
          </div>
        </div>
        <div className="flex flex-row gap-4">
          <Button
            variant="secondary"
            disabled={!item}
            onClick={() => onCancel()}
          >
            <ButtonText text="Cancel" />
          </Button>
          <Button disabled={!item} onClick={() => onSave()}>
            <ButtonText text="Save Row" />
          </Button>
        </div>
      </div>

      <div>
        <Textarea
          disabled={!item}
          placeholder={
            !!item
              ? 'Add a table filter here'
              : 'Click edit on a row above to change the where clause'
          }
          value={item?.where ?? ''}
          onChange={(e) => onWhereChange(e.currentTarget.value)}
        />
      </div>
      <div>
        <Textarea disabled={true} value={buildSelectQuery(item?.where)} />
      </div>
    </div>
  );
}

function buildSelectQuery(whereClause?: string): string {
  if (!whereClause) {
    return '';
  }
  return `WHERE ${whereClause};`;
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
    (c) => c.id == formData.flow.sourceId
  );
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
        transformer: toTransformerConfigOptions(m.transformer),
      });
    }),
    source: new JobSource({
      connectionId: formData.flow.sourceId,
      options: toJobSourceOptions(formData, sourceConnection),
    }),
    destinations: formData.flow.destinations.map((d) => {
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
            case: 'postgresOptions',
            value: new PostgresSourceConnectionOptions({
              haltOnNewColumnAddition:
                values.flow.sourceOptions.haltOnNewColumnAddition,
            }),
          },
        });
      case 'mysqlConfig':
        return new JobSourceOptions({
          config: {
            case: 'mysqlOptions',
            value: new MysqlSourceConnectionOptions({
              haltOnNewColumnAddition:
                values.flow.sourceOptions.haltOnNewColumnAddition,
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
