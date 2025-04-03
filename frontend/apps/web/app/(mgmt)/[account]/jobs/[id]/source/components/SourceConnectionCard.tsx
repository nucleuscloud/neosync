'use client';
import { useQuery } from '@connectrpc/connect-query';
import { JobService } from '@neosync/sdk';
import { ReactElement } from 'react';
import { isAiDataGenJob, isDataGenJob, isPiiDetectJob } from '../../util';
import AiDataGenConnectionCard from './AiDataGenConnectionCard';
import DataGenConnectionCard from './DataGenConnectionCard';
import DataSyncConnectionCard from './DataSyncConnectionCard';
import PiiDetectConnectionCard from './PiiDetectConnectionCard';
import SchemaPageSkeleton from './SchemaPageSkeleton';

interface Props {
  jobId: string;
}

export default function SourceConnectionCard({ jobId }: Props): ReactElement {
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
  if (isPiiDetectJob(data?.job)) {
    return <PiiDetectConnectionCard jobId={jobId} />;
  }
  return <DataSyncConnectionCard jobId={jobId} />;
}
