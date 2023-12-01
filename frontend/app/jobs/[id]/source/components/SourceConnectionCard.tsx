'use client';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { JobSource } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { ReactElement } from 'react';
import { isDataGenJob } from '../../util';
import DataSyncConnectionCard from './DataSyncConnectionCard';

interface Props {
  jobId: string;
}

export function getConnectionIdFromSource(
  js: JobSource | undefined
): string | undefined {
  if (
    js?.options?.config.case === 'postgres' ||
    js?.options?.config.case === 'mysql' ||
    js?.options?.config.case === 'awsS3'
  ) {
    return js.options.config.value.connectionId;
  }
  return undefined;
}

export default function SourceConnectionCard({ jobId }: Props): ReactElement {
  const { data, isLoading } = useGetJob(jobId);

  if (isLoading) {
    return (
      <div className="space-y-10">
        <Skeleton className="w-full h-12" />
        <Skeleton className="w-1/2 h-12" />
        <SkeletonTable />
      </div>
    );
  }
  if (isDataGenJob(data?.job)) {
    return <div>todo</div>;
  }
  return <DataSyncConnectionCard jobId={jobId} />;
}
