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
import { SchemaFormValues } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import { SUBSET_FORM_SCHEMA, SubsetFormValues } from '../schema';
import { TableRow, getColumns } from './schema-table/column';
import { DataTable } from './schema-table/data-table';

export default function Page({ searchParams }: PageProps): ReactElement {
  const account = useAccount();
  const router = useRouter();
  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = searchParams?.sessionId ?? '';
  const sessionKey = `${sessionPrefix}-new-job-subset`;

  const [subsetFormValues] = useSessionStorage<SubsetFormValues>(sessionKey, {
    subsets: [],
  });

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

  const [schemaFormValues] = useSessionStorage<SchemaFormValues>(
    `${sessionPrefix}-new-job-schema`,
    {
      mappings: [],
    }
  );

  const [itemToEdit, setItemToEdit] = useState<TableRow | undefined>();

  async function onSubmit(values: SubsetFormValues): Promise<void> {}

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
              <SchemaTable
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
            <div className="my-6">
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

interface SchemaTableProps {
  data: TableRow[];
  onEdit(schema: string, table: string): void;
}

function SchemaTable(props: SchemaTableProps): ReactElement {
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
          placeholder="Add a table filter here"
          value={item?.where ?? ''}
          onChange={(e) => onWhereChange(e.currentTarget.value)}
        />
      </div>
    </div>
  );
}
