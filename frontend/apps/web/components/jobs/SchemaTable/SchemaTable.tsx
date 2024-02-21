'use client';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { useGetMergedTransformers } from '@/libs/hooks/useGetMergedTransformers';
import { JobMappingFormValues } from '@/yup-validations/jobs';
import {
  ForeignConstraintTables,
  GetConnectionSchemaResponse,
  PrimaryConstraint,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { getSchemaColumns } from './SchemaColumns';
import SchemaPageTable, { Row } from './SchemaPageTable';

interface Props {
  data: JobMappingFormValues[];
  excludeInputReqTransformers?: boolean; // will result in only generators (functions with no data input)
  primaryConstraints?: Record<string, PrimaryConstraint>;
  foreignConstraints?: Record<string, ForeignConstraintTables>;
}

export function SchemaTable(props: Props): ReactElement {
  const {
    data,
    excludeInputReqTransformers,
    primaryConstraints,
    foreignConstraints,
  } = props;

  const { account } = useAccount();
  const { mergedTransformers, isLoading } = useGetMergedTransformers(
    excludeInputReqTransformers ?? false,
    account?.id ?? ''
  );

  const tableData = data.map((d, idx): Row => {
    return {
      ...d,
      formIdx: idx, // this is very important because we need to retain this when updating the form due to the table being filterable. Otherwise the index used is incorrect.
    };
  });

  const columns = getSchemaColumns({
    transformers: mergedTransformers,
  });

  if (isLoading || !tableData || tableData.length == 0) {
    return <SkeletonTable />;
  }

  // tie in constraint data to the rows so we can display and filter by them
  data.forEach((row: JobMappingFormValues) => {
    // build map look up key
    const schemaTable = `${row.schema}.${row.table}`;

    // add primary key constraints field so that we can filter by it in the table
    // set it to string literal 'Primary Key' so that's what gets rendered in the table
    if (primaryConstraints?.[schemaTable]?.columns.includes(row.column)) {
      row.primaryConstraints = 'Primary Key';
    }

    // check to see if foreign key exists and if it does then add it to the row
    if (
      foreignConstraints &&
      foreignConstraints[schemaTable]?.constraints.filter(
        (item) => item.column == row.column
      ).length > 0
    ) {
      const fk = foreignConstraints[schemaTable]?.constraints;
      row.foreignConstraints = {
        table: fk[0].foreignKey?.table ?? '', // the foreignKey constraints object comes back from the API with two identical objects in an array, so just getting the first one. Need to investigate why it returns two.
        column: fk[0].foreignKey?.column ?? '',
        value: 'Foreign Key',
      };
    }
  });

  return (
    <div>
      <SchemaPageTable
        columns={columns}
        data={tableData}
        transformers={mergedTransformers}
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
