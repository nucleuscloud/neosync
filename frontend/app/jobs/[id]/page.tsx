'use client';
import { PageProps } from '@/components/types';
import { useGetJob } from '@/libs/hooks/useGetJob';

import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { useGetJobStatus } from '@/libs/hooks/useGetJobStatus';
import { ReactElement } from 'react';
import JobNextRuns from './components/NextRuns';
import JobPauseSwitch from './components/PauseSwitch';
import JobRecentRuns from './components/RecentRuns';
import JobScheduleCard from './components/ScheduleCard';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data, isLoading, mutate } = useGetJob(id);
  const { data: jobStatus } = useGetJobStatus(id);

  if (isLoading) {
    return (
      <div className="mt-10">
        <SkeletonForm />
      </div>
    );
  }

  if (!data?.job) {
    return (
      <div className="mt-10">
        <Alert variant="destructive">
          <AlertTitle>{`Error: Unable to retrieve job`}</AlertTitle>
        </Alert>
      </div>
    );
  }

  return (
    <div className="job-details-container">
      <div className="space-y-10">
        <JobPauseSwitch jobId={id} status={jobStatus?.status} mutate={mutate} />
        <JobScheduleCard job={data?.job} mutate={mutate} />
        <JobRecentRuns jobId={id} />
        <div className="flex">
          <JobNextRuns jobId={id} status={jobStatus?.status} />
        </div>
      </div>
    </div>
  );
}
