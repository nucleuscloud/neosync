'use client';
import {
  filterInputFreeSystemTransformers,
  filterInputFreeUdfTransformers,
} from '@/app/transformers/EditTransformerOptions';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { useGetSystemTransformers } from '@/libs/hooks/useGetSystemTransformers';
import { useGetUserDefinedTransformers } from '@/libs/hooks/useGetUserDefinedTransformers';
import { GetConnectionSchemaResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { joinTransformers } from '@/shared/transformers';
import { JobMappingFormValues } from '@/yup-validations/jobs';
import { ReactElement } from 'react';
import { VirtualizedSchemaTable } from './VirtualizedSchemaTable';

interface Props {
  data?: JobMappingFormValues[];
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

  const tableData = data?.map((d) => {
    return {
      ...d,
      isSelected: false,
    };
  });

  if (
    systemTransformersIsLoading ||
    customTransformersIsLoading ||
    !tableData
  ) {
    return <SkeletonTable />;
  }

  return (
    <div>
      <VirtualizedSchemaTable
        data={tableData}
        transformers={mergedTransformers}
      />
    </div>
  );
}

export async function getConnectionSchema(
  connectionId?: string
): Promise<GetConnectionSchemaResponse | undefined> {
  if (!connectionId) {
    return;
  }
  const res = await fetch(`/api/connections/${connectionId}/schema`, {
    method: 'GET',
    headers: {
      'content-type': 'application/json',
    },
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return GetConnectionSchemaResponse.fromJson(await res.json());
}
