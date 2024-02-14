'use client';
import {
  filterInputFreeSystemTransformers,
  filterInputFreeUdfTransformers,
} from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { useGetSystemTransformers } from '@/libs/hooks/useGetSystemTransformers';
import { useGetUserDefinedTransformers } from '@/libs/hooks/useGetUserDefinedTransformers';
import { joinTransformers } from '@/shared/transformers';
import { JobMappingFormValues } from '@/yup-validations/jobs';
import { GetConnectionSchemaResponse } from '@neosync/sdk';
import { ReactElement, useMemo } from 'react';
import { getSchemaColumns } from './SchemaColumns';
import SchemaPageTable, { Row } from './SchemaPageTable';

interface Props {
  data: JobMappingFormValues[];
  excludeInputReqTransformers?: boolean; // will result in only generators (functions with no data input)
}

export function SchemaTable(props: Props): ReactElement {
  const { data, excludeInputReqTransformers } = props;

  const { account } = useAccount();
  const { data: systemTransformers, isLoading: systemTransformersIsLoading } =
    useGetSystemTransformers();
  const { data: customTransformers, isLoading: customTransformersIsLoading } =
    useGetUserDefinedTransformers(account?.id ?? '');

  const filteredSystemTransformers = excludeInputReqTransformers
    ? filterInputFreeSystemTransformers(systemTransformers?.transformers ?? [])
    : systemTransformers?.transformers ?? [];

  const filteredCustomTransformers = excludeInputReqTransformers
    ? filterInputFreeUdfTransformers(
        customTransformers?.transformers ?? [],
        filteredSystemTransformers
      )
    : customTransformers?.transformers ?? [];

  const mergedTransformers = joinTransformers(
    filteredSystemTransformers,
    filteredCustomTransformers
  );

  const tableData = data.map((d, idx): Row => {
    return {
      ...d,
      formIdx: idx, // this is very important because we need to retain this when updating the form due to the table being filterable. Otherwise the index used is incorrect.
    };
  });

  const columns = useMemo(
    () => getSchemaColumns({ transformers: mergedTransformers }),
    [mergedTransformers] // Use mergedTransformers instead of transformers
  );

  if (
    systemTransformersIsLoading ||
    customTransformersIsLoading ||
    !tableData ||
    tableData.length == 0
  ) {
    return <SkeletonTable />;
  }

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
