import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { useGetTransformers } from '@/libs/hooks/useGetTransformers';
import { GetConnectionSchemaResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { JobMappingFormValues } from '@/yup-validations/jobs';
import { ReactElement } from 'react';
import { VirtualizedSchemaTable } from './VirtualizedSchemaTable';

interface JobTableProps {
  data: JobMappingFormValues[];
}

export function SchemaTable(props: JobTableProps): ReactElement {
  const { data } = props;
  const { data: transformers, isLoading: transformersIsLoading } =
    useGetTransformers();

  if (transformersIsLoading) {
    return <SkeletonTable />;
  }

  const tableData = data.map((d) => {
    return {
      ...d,
      isSelected: false,
    };
  });
  // const tableData = makeData(10000);

  return (
    <div>
      <VirtualizedSchemaTable
        data={tableData}
        transformers={transformers?.transformers}
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
