'use client';
import { ColumnMetadata } from '@/app/(mgmt)/[account]/new/job/schema/page';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { useGetMergedTransformers } from '@/libs/hooks/useGetMergedTransformers';
import { JobMappingFormValues } from '@/yup-validations/jobs';
import { GetConnectionSchemaResponse } from '@neosync/sdk';
import { ReactElement } from 'react';
import { getSchemaColumns } from './SchemaColumns';
import SchemaPageTable, { Row } from './SchemaPageTable';

interface Props {
  data: JobMappingFormValues[];
  excludeInputReqTransformers?: boolean; // will result in only generators (functions with no data input)
  columnMetadata: ColumnMetadata;
  jobType: string;
}

export function SchemaTable(props: Props): ReactElement {
  const { data, excludeInputReqTransformers, columnMetadata, jobType } = props;

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
    columnMetadata: columnMetadata,
  });

  if (isLoading || !tableData || tableData.length == 0) {
    return <SkeletonTable />;
  }

  return (
    <div>
      <SchemaPageTable
        columns={columns}
        data={tableData}
        transformers={mergedTransformers}
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
