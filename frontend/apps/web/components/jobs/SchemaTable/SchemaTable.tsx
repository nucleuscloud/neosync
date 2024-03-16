'use client';
import DualListBox, { Action } from '@/components/DualListBox/DualListBox';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Card, CardContent } from '@/components/ui/card';
import { ConnectionSchemaMap } from '@/libs/hooks/useGetConnectionSchemaMap';
import { useGetMergedTransformers } from '@/libs/hooks/useGetMergedTransformers';
import {
  JobMappingFormValues,
  SchemaFormValues,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import {
  GetConnectionSchemaResponse,
  JobMappingTransformer,
  Passthrough,
  TransformerConfig,
} from '@neosync/sdk';
import { ReactElement, useState } from 'react';
import { useFieldArray, useFormContext } from 'react-hook-form';
import { SchemaConstraintHandler, getSchemaColumns } from './SchemaColumns';
import SchemaPageTable, { Row } from './SchemaPageTable';

interface Props {
  data: JobMappingFormValues[];
  excludeInputReqTransformers?: boolean; // will result in only generators (functions with no data input)
  jobType: string; // todo: update to be named type
  schema: ConnectionSchemaMap;
  constraintHandler: SchemaConstraintHandler;
}

export function SchemaTable(props: Props): ReactElement {
  const {
    data,
    excludeInputReqTransformers,
    constraintHandler,
    jobType,
    schema,
  } = props;

  const { account } = useAccount();
  const { transformers, isLoading } = useGetMergedTransformers(
    excludeInputReqTransformers ?? false,
    account?.id ?? ''
  );
  const [selectedItems, setSelectedItems] = useState<Set<string>>(
    new Set(data.map((d) => `${d.schema}.${d.table}`))
  );

  const tableData = data.map((d, idx): Row => {
    return {
      ...d,
      formIdx: idx, // this is very important because we need to retain this when updating the form due to the table being filterable. Otherwise the index used is incorrect.
    };
  });

  const columns = getSchemaColumns({
    transformers,
    constraintHandler,
  });

  const form = useFormContext<SchemaFormValues>();
  const { append, remove, fields } = useFieldArray<SchemaFormValues>({
    control: form.control,
    name: 'mappings',
  });

  function toggleItem(items: Set<string>, action: Action): void {
    if (items.size === 0) {
      const idxs = fields.map((_, idx) => idx);
      remove(idxs);
      setSelectedItems(new Set());
      return;
    }
    const [added, removed] = getDelta(items, selectedItems);

    const toRemove: number[] = [];
    const toAdd: any[] = [];

    fields.forEach((field, idx) => {
      if (removed.has(`${field.schema}.${field.table}`)) {
        toRemove.push(idx);
      }
    });
    added.forEach((item) => {
      const dbcols = schema[item];
      if (!dbcols) {
        return;
      }
      dbcols.forEach((dbcol) => {
        toAdd.push({
          schema: dbcol.schema,
          table: dbcol.table,
          column: dbcol.column,
          dataType: dbcol.dataType,
          transformer: convertJobMappingTransformerToForm(
            new JobMappingTransformer({
              source: 'passthrough',
              config: new TransformerConfig({
                config: {
                  case: 'passthroughConfig',
                  value: new Passthrough({}),
                },
              }),
            })
          ),
        });
      });
    });
    if (toRemove.length > 0) {
      remove(toRemove);
    }
    if (toAdd.length > 0) {
      append(toAdd);
    }
    setSelectedItems(items);
  }

  if (isLoading || !tableData) {
    return <SkeletonTable />;
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex">
        <Card className="p-0">
          <CardContent className="p-3">
            <DualListBox
              options={Object.keys(schema).map((value) => ({
                value: value,
              }))}
              selected={selectedItems}
              onChange={toggleItem}
              title="Table"
            />
          </CardContent>
        </Card>
      </div>
      <div></div>
      <SchemaPageTable
        columns={columns}
        data={tableData}
        transformers={transformers}
        jobType={jobType}
      />
    </div>
  );
}

export async function getConnectionSchema(
  accountId: string,
  connectionId?: string
): Promise<GetConnectionSchemaResponse | undefined> {
  if (!accountId || !connectionId) {
    return;
  }
  const res = await fetch(
    `/api/accounts/${accountId}/connections/${connectionId}/schema`,
    {
      method: 'GET',
      headers: {
        'content-type': 'application/json',
      },
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return GetConnectionSchemaResponse.fromJson(await res.json());
}

function getDelta(
  newSet: Set<string>,
  oldSet: Set<string>
): [Set<string>, Set<string>] {
  const added = new Set<string>();
  const removed = new Set<string>();

  oldSet.forEach((val) => {
    if (!newSet.has(val)) {
      removed.add(val);
    }
  });
  newSet.forEach((val) => {
    if (!oldSet.has(val)) {
      added.add(val);
    }
  });

  return [added, removed];
}
