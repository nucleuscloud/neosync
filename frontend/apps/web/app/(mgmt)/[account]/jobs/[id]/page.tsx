'use client';
import { PageProps } from '@/components/types';
import { useGetJob } from '@/libs/hooks/useGetJob';

import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { useGetJobStatus } from '@/libs/hooks/useGetJobStatus';
import { GetJobResponse } from '@neosync/sdk';
import { ReactElement } from 'react';
import JobNextRuns from './components/NextRuns';
import JobRecentRuns from './components/RecentRuns';
import JobScheduleCard from './components/ScheduleCard';
import WorkflowSettingsCard from './components/WorkflowSettingsCard';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { account } = useAccount();
  const { data, isLoading, mutate } = useGetJob(account?.id ?? '', id);
  const { data: jobStatus } = useGetJobStatus(account?.id ?? '', id);

  if (isLoading) {
    return (
      <div className="pt-10">
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
      <div className="flex flex-col gap-5">
        <div className="flex flex-row gap-5">
          <div className="flex-grow basis-3/4">
            <JobScheduleCard job={data.job} mutate={mutate} />
          </div>
          <div className="flex-grow basis-1/4 overflow-y-auto rounded-xl border border-card-border">
            <JobNextRuns jobId={id} status={jobStatus?.status} />
          </div>
        </div>
        <JobRecentRuns jobId={id} />
        <div className="flex">
          <WorkflowSettingsCard
            job={data.job}
            mutate={(newjob) => mutate(new GetJobResponse({ job: newjob }))}
          />
        </div>
      </div>
    </div>
  );
}
