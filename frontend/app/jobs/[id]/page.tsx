'use client';
import { PageProps } from '@/components/types';
import { useGetJob } from '@/libs/hooks/useGetJob';

import SubPageHeader from '@/components/headers/SubPageHeader';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import { useGetJobStatus } from '@/libs/hooks/useGetJobStatus';
import { getErrorMessage } from '@/util/util';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import JobNextRuns from './components/NextRuns';
import JobPauseSwitch from './components/PauseSwitch';
import JobRecentRuns from './components/RecentRuns';
import JobScheduleCard from './components/ScheduleCard';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data, isLoading, mutate } = useGetJob(id);
  const { data: jobStatus } = useGetJobStatus(id);
  const router = useRouter();
  const { toast } = useToast();

  async function onTriggerJobRun(): Promise<void> {
    try {
      await triggerJobRun(id);
      toast({
        title: 'Job run triggered successfully!',
      });
      router.push(`/runs`);
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to trigger job run',
        description: getErrorMessage(err),
      });
    }
  }

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
        extraHeading={
          <Button onClick={() => onTriggerJobRun()}>Trigger Run</Button>
        }
      />
      <div className="space-y-10">
        <JobPauseSwitch jobId={id} status={jobStatus?.status} mutate={mutate} />
        <JobScheduleCard job={data?.job} mutate={mutate} />
        <div className="flex md:flex-row md:space-x-8 md:space-y-0 flex-col space-y-8">
          <div className="md:basis-1/4">
            <JobNextRuns jobId={id} jobStatus={jobStatus?.status} />
          </div>
          <div className="md:basis-3/4">
            <JobRecentRuns jobId={id} />
          </div>
        </div>
      </div>
    </div>
  );
}

async function triggerJobRun(jobId: string): Promise<void> {
  const res = await fetch(`/api/jobs/${jobId}/create-run`, {
    method: 'POST',
    body: JSON.stringify({ jobId }),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}
