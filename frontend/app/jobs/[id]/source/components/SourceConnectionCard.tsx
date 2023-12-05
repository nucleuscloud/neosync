'use client';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { ReactElement } from 'react';
import { isDataGenJob } from '../../util';
import DataGenConnectionCard from './DataGenConnectionCard';
import DataSyncConnectionCard from './DataSyncConnectionCard';

interface Props {
  jobId: string;
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
    return <DataGenConnectionCard jobId={jobId} />;
  }
  return <DataSyncConnectionCard jobId={jobId} />;
}
