'use client';
import { PageProps } from '@/components/types';
import { useGetJob } from '@/libs/hooks/useGetJob';

import SubPageHeader from '@/components/headers/SubPageHeader';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { ReactElement } from 'react';
import JobNextRuns from './components/NextRuns';
import JobPauseSwitch from './components/PauseSwitch';
import JobRecentRuns from './components/RecentRuns';
import JobScheduleCard from './components/ScheduleCard';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data, isLoading, mutate } = useGetJob(id);

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
      <SubPageHeader
        header={data?.job?.name || ''}
        description={data?.job?.id || ''}
      />
      <div className="space-y-10">
        <JobPauseSwitch job={data?.job} mutate={mutate} />
        <JobScheduleCard job={data?.job} mutate={mutate} />
        <div className="flex flex-row space-x-8">
          <div className="basis-1/4">
            <JobNextRuns job={data?.job} />
          </div>
          <div className="basis-3/4">
            <JobRecentRuns job={data?.job} />
          </div>
        </div>
      </div>
    </div>
  );
}
