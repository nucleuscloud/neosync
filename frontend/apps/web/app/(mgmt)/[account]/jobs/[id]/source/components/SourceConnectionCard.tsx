'use client';
import { useAccount } from '@/components/providers/account-provider';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { ReactElement } from 'react';
import { isAiDataGenJob, isDataGenJob } from '../../util';
import AiDataGenConnectionCard from './AiDataGenConnectionCard';
import DataGenConnectionCard from './DataGenConnectionCard';
import DataSyncConnectionCard from './DataSyncConnectionCard';
import SchemaPageSkeleton from './SchemaPageSkeleton';

interface Props {
  jobId: string;
}

export default function SourceConnectionCard({ jobId }: Props): ReactElement {
  const { account } = useAccount();
  const { data, isLoading } = useGetJob(account?.id ?? '', jobId);

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
