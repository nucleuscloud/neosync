import { MergeSystemAndCustomTransformers } from '@/app/transformers/EditTransformerOptions';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { useGetCustomTransformers } from '@/libs/hooks/useGetCustomTransformers';
import { useGetSystemTransformers } from '@/libs/hooks/useGetSystemTransformers';
import { GetConnectionSchemaResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { JobMappingFormValues } from '@/yup-validations/jobs';
import { ReactElement } from 'react';
import { VirtualizedSchemaTable } from './VirtualizedSchemaTable';

interface JobTableProps {
  data?: JobMappingFormValues[];
}

export function SchemaTable(props: JobTableProps): ReactElement {
  const { data } = props;

  const { account } = useAccount();
  const { data: systemTransformers, isLoading: systemTransformersIsLoading } =
    useGetSystemTransformers();

  const { data: customTransformers, isLoading: customTransformersIsLoading } =
    useGetCustomTransformers(account?.id ?? '');

  const mergedTransformers = MergeSystemAndCustomTransformers(
    systemTransformers?.transformers ?? [],
    customTransformers?.transformers ?? []
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

  console.log('merged', mergedTransformers);

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
