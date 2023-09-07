import { getColumns } from '@/app/jobs/components/SchemaTable/column';
import { DataTable } from '@/app/jobs/components/SchemaTable/data-table';
import { JobMappingFormValues } from '@/app/new/job/schema';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetTransformers } from '@/libs/hooks/useGetTransformers';
import { GetConnectionSchemaResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { ReactElement } from 'react';

export interface JobTableProps {
  data: JobMappingFormValues[];
}

export function SchemaTable(props: JobTableProps): ReactElement {
  const { data } = props;
  const { data: transformers, isLoading: transformersIsLoading } =
    useGetTransformers();

  if (transformersIsLoading) {
    return <Skeleton />;
  }

  const columns = getColumns({ transformers: transformers?.transformers });

  return (
    <div>
      <DataTable columns={columns} data={data} />
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
