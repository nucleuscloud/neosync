'use client';
import { PageProps } from '@/components/types';
import { useGetJob } from '@/libs/hooks/useGetJob';

import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { useGetJobStatus } from '@/libs/hooks/useGetJobStatus';
import { GetJobResponse } from '@neosync/sdk';
import { ReactElement } from 'react';
import ActivitySyncOptionsCard from './components/ActivitySyncOptionsCard';
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
            <JobScheduleCard
              job={data.job}
              mutate={(newjob) => mutate(new GetJobResponse({ job: newjob }))}
            />
          </div>
          <div className="flex-grow basis-1/4 overflow-y-auto rounded-xl border border-card-border">
            <JobNextRuns jobId={id} status={jobStatus?.status} />
          </div>
        </div>
        <JobRecentRuns jobId={id} />
        <Accordion type="single" collapsible>
          <AccordionItem value="advanced-settings">
            <AccordionTrigger className="hover:no-underline">
              <div className="flex flex-col gap-1 px-1 items-center">
                <h2 className="text-2xl font-semibold">Advanced Settings</h2>
                <p className="text-xs text-muted-foreground">
                  Click to see more advanced options
                </p>
              </div>
            </AccordionTrigger>
            <AccordionContent>
              <div className="flex flex-col gap-3">
                <div>
                  <WorkflowSettingsCard
                    job={data.job}
                    mutate={(newjob) =>
                      mutate(new GetJobResponse({ job: newjob }))
                    }
                  />
                </div>
                <div>
                  <ActivitySyncOptionsCard
                    job={data.job}
                    mutate={(newjob) =>
                      mutate(new GetJobResponse({ job: newjob }))
                    }
                  />
                </div>
              </div>
            </AccordionContent>
          </AccordionItem>
        </Accordion>
        {/* <div className="flex flex-col gap-3">
          <div>
            <h2 className="text-2xl font-semibold tracking-tight px-1">
              Advanced Settings
            </h2>
          </div>
          <div className="flex flex-col gap-3">
            <div>
              <WorkflowSettingsCard
                job={data.job}
                mutate={(newjob) => mutate(new GetJobResponse({ job: newjob }))}
              />
            </div>
            <div>
              <ActivitySyncOptionsCard
                job={data.job}
                mutate={(newjob) => mutate(new GetJobResponse({ job: newjob }))}
              />
            </div>
          </div>
        </div> */}
      </div>
    </div>
  );
}
