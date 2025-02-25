'use client';
import { useQuery } from '@connectrpc/connect-query';
import { JobService } from '@neosync/sdk';
import { ReactElement } from 'react';
import { isAiDataGenJob, isDataGenJob } from '../../util';
import AiDataGenConnectionCard from './AiDataGenConnectionCard';
import DataGenConnectionCard from './DataGenConnectionCard';
import DataSyncConnectionCard from './DataSyncConnectionCard';
import SchemaPageSkeleton from './SchemaPageSkeleton';

interface Props {
  jobId: string;
}

export default function SourceConnectionCard({ jobId }: Props): ReactElement<any> {
  const { data, isLoading } = useQuery(
    JobService.method.getJob,
    { id: jobId },
    { enabled: !!jobId }
  );

  if (isLoading) {
    return <SchemaPageSkeleton />;
  }
  if (isDataGenJob(data?.job)) {
    return <DataGenConnectionCard jobId={jobId} />;
  }
  if (isAiDataGenJob(data?.job)) {
    return <AiDataGenConnectionCard jobId={jobId} />;
  }
  return <DataSyncConnectionCard jobId={jobId} />;
}
