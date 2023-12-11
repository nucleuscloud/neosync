import {
  SUBSET_FORM_SCHEMA,
  SubsetFormValues,
} from '@/app/[account]/new/job/schema';
import EditItem from '@/components/jobs/subsets/EditItem';
import SubsetTable from '@/components/jobs/subsets/subset-table/SubsetTable';
import { TableRow } from '@/components/jobs/subsets/subset-table/column';
import {
  buildRowKey,
  buildTableRowData,
} from '@/components/jobs/subsets/utils';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { useToast } from '@/components/ui/use-toast';
import { useGetConnectionSchema } from '@/libs/hooks/useGetConnectionSchema';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { getErrorMessage } from '@/util/util';
import { toPostgresSourceSchemaOptions } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  GetJobResponse,
  JobSourceOptions,
  JobSourceSqlSubetSchemas,
  PostgresSourceSchemaSubset,
  SetJobSourceSqlConnectionSubsetsRequest,
  SetJobSourceSqlConnectionSubsetsResponse,
} from '@neosync/sdk';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';
import { getConnectionIdFromSource } from '../../source/components/util';

interface Props {
  jobId: string;
}

export default function SubsetCard(props: Props): ReactElement {
  const { jobId } = props;
  const { toast } = useToast();
  const { account } = useAccount();
  const {
    data,
    mutate: mutateJob,
    isLoading: isJobLoading,
  } = useGetJob(account?.id ?? '', jobId);
  const sourceConnectionId = getConnectionIdFromSource(data?.job?.source);
  const { data: schema, isLoading: isSchemaLoading } = useGetConnectionSchema(
    account?.id ?? '',
    sourceConnectionId
  );

  const formValues = getFormValues(data?.job?.source?.options);
  const form = useForm({
    resolver: yupResolver<SubsetFormValues>(SUBSET_FORM_SCHEMA),
    defaultValues: { subsets: [] },
    values: formValues,
  });

  const tableRowData = buildTableRowData(
    schema?.schemas ?? [],
    form.watch().subsets // ensures that all form changes cause a re-render since stuff happens outside of the form that depends on the form values
  );
  const [itemToEdit, setItemToEdit] = useState<TableRow | undefined>();

  const formValuesMap = new Map(
    formValues.subsets.map((ss) => [buildRowKey(ss.schema, ss.table), ss])
  );

  if (isJobLoading || isSchemaLoading) {
    return (
      <div className="space-y-10">
        <Skeleton className="w-full h-12" />
        <Skeleton className="w-1/2 h-12" />
        <SkeletonTable />
      </div>
    );
  }

  async function onSubmit(values: SubsetFormValues): Promise<void> {
    try {
      const updatedJobRes = await setJobSubsets(
        account?.id ?? '',
        jobId,
        values
      );
      toast({
        title: 'Successfully updated database subsets',
        variant: 'success',
      });
      mutateJob(
        new GetJobResponse({
          job: updatedJobRes.job,
        })
      );
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to update database subsets',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  function hasLocalChange(schema: string, table: string): boolean {
    const key = buildRowKey(schema, table);
    const trData = tableRowData[key];

    const svrData = formValuesMap.get(key);

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
      const svrData = formValuesMap.get(key);
      form.setValue(`subsets.${idx}`, {
        schema: schema,
        table: table,
        whereClause: svrData?.whereClause ?? undefined,
      });
    }
  }
  return (
    <div>
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
              connectionId={sourceConnectionId ?? ''}
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
          <div className="flex flex-col gap-2">
            <div className="flex flex-row justify-end">
              <p className="text-sm tracking-tight">
                Save changes to apply any updates made to the table above
              </p>
            </div>
            <div className="flex flex-row gap-1 justify-end">
              <Button key="submit" type="submit">
                Save
              </Button>
            </div>
          </div>
        </form>
      </Form>
    </div>
  );
}

function getFormValues(sourceOpts?: JobSourceOptions): SubsetFormValues {
  if (
    !sourceOpts ||
    (sourceOpts.config.case !== 'postgres' &&
      sourceOpts.config.case !== 'mysql')
  ) {
    return { subsets: [] };
  }

  const schemas = sourceOpts.config.value.schemas;
  const subsets: SubsetFormValues['subsets'] = schemas.flatMap((schema) => {
    return schema.tables.map((table) => {
      return {
        schema: schema.schema,
        table: table.table,
        whereClause: table.whereClause,
      };
    });
  });
  return { subsets };
}

async function setJobSubsets(
  accountId: string,
  jobId: string,
  values: SubsetFormValues
): Promise<SetJobSourceSqlConnectionSubsetsResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${jobId}/source-connection/subsets`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(
        new SetJobSourceSqlConnectionSubsetsRequest({
          id: jobId,
          schemas: new JobSourceSqlSubetSchemas({
            schemas: {
              case: 'postgresSubset',
              value: new PostgresSourceSchemaSubset({
                postgresSchemas: toPostgresSourceSchemaOptions(values.subsets),
              }),
            },
          }),
        })
      ),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return SetJobSourceSqlConnectionSubsetsResponse.fromJson(await res.json());
}
