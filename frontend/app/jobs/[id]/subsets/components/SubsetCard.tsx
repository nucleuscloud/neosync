import { SUBSET_FORM_SCHEMA, SubsetFormValues } from '@/app/new/job/schema';
import { TableRow, getColumns } from '@/app/new/job/subset/subset-table/column';
import { DataTable } from '@/app/new/job/subset/subset-table/data-table';
import ButtonText from '@/components/ButtonText';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { Textarea } from '@/components/ui/textarea';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { useToast } from '@/components/ui/use-toast';
import { useGetConnectionSchema } from '@/libs/hooks/useGetConnectionSchema';
import { useGetJob } from '@/libs/hooks/useGetJob';
import {
  GetJobResponse,
  JobSourceOptions,
  JobSourceSqlSubetSchemas,
  PostgresSourceSchemaSubset,
  SetJobSourceSqlConnectionSubsetsRequest,
  SetJobSourceSqlConnectionSubsetsResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import { toPostgresSourceSchemaOptions } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';

interface Props {
  jobId: string;
}

function getFormValues(sourceOpts?: JobSourceOptions): SubsetFormValues {
  if (
    !sourceOpts ||
    (sourceOpts.config.case !== 'mysqlOptions' &&
      sourceOpts.config.case !== 'postgresOptions')
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

export default function SubsetCard(props: Props): ReactElement {
  const { jobId } = props;
  const { toast } = useToast();
  const { data, mutate: mutateJob, isLoading: isJobLoading } = useGetJob(jobId);
  const { data: schema, isLoading: isSchemaLoading } = useGetConnectionSchema(
    data?.job?.source?.connectionId
  );

  const form = useForm({
    resolver: yupResolver<SubsetFormValues>(SUBSET_FORM_SCHEMA),
    defaultValues: { subsets: [] },
    values: getFormValues(data?.job?.source?.options),
  });

  const tableRowData = buildTableRowData(
    schema?.schemas ?? [],
    form.getValues().subsets
  );
  const [itemToEdit, setItemToEdit] = useState<TableRow | undefined>();

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
      const updatedJobRes = await setJobSubsets(jobId, values);
      toast({
        title: 'Successfully updated database subsets',
        variant: 'default',
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
  console.log('errors', form.formState.errors);
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
                      buildRowKey(itemToEdit.schema, itemToEdit.table)
                  );
                if (idx >= 0) {
                  console.log('editing idx subsets', idx, itemToEdit);
                  form.setValue(`subsets.${idx}`, {
                    schema: itemToEdit.schema,
                    table: itemToEdit.table,
                    whereClause: itemToEdit.where,
                  });
                } else {
                  console.log('appending subsets', itemToEdit);
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

///////

interface SubsetTableProps {
  data: TableRow[];
  onEdit(schema: string, table: string): void;
}

function SubsetTable(props: SubsetTableProps): ReactElement {
  const { data, onEdit } = props;

  const columns = getColumns({ onEdit });

  return <DataTable columns={columns} data={data} />;
}

interface DbCol {
  schema: string;
  table: string;
}
function buildTableRowData(
  dbCols: DbCol[],
  existingSubsets: SubsetFormValues['subsets']
): Record<string, TableRow> {
  const tableMap: Record<string, TableRow> = {};

  dbCols.forEach((mapping) => {
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
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button disabled={!item} onClick={() => onSave()}>
                  <ButtonText text="Apply" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>
                  Applies changes to table only, click Save below to fully
                  submit changes
                </p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
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

async function setJobSubsets(
  jobId: string,
  values: SubsetFormValues
): Promise<SetJobSourceSqlConnectionSubsetsResponse> {
  const res = await fetch(`/api/jobs/${jobId}/source-connection/subsets`, {
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
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return SetJobSourceSqlConnectionSubsetsResponse.fromJson(await res.json());
}
