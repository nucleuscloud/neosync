'use client';
import { PageProps } from '@/components/types';

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { createConnectQueryKey, useQuery } from '@connectrpc/connect-query';
import { GetJobResponse } from '@neosync/sdk';
import { getJob, getJobStatus } from '@neosync/sdk/connectquery';
import { useQueryClient } from '@tanstack/react-query';
import { ReactElement } from 'react';
import ActivitySyncOptionsCard from './components/ActivitySyncOptionsCard';
import JobNextRuns from './components/NextRuns';
import JobRecentRuns from './components/RecentRuns';
import JobScheduleCard from './components/ScheduleCard';
import WorkflowSettingsCard from './components/WorkflowSettingsCard';
import JobIdSkeletonForm from './JobIdSkeletonForm';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data, isLoading } = useQuery(getJob, { id }, { enabled: !!id });
  const { data: jobStatus } = useQuery(getJobStatus, { jobId: id });
  const queryclient = useQueryClient();

  if (isLoading) {
    return (
      <div className="pt-10">
        <JobIdSkeletonForm />
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
              mutate={(newjob) => {
                queryclient.setQueryData(
                  createConnectQueryKey(getJob, { id }),
                  new GetJobResponse({ job: newjob })
                );
              }}
            />
          </div>
          <div className="flex-grow basis-1/4 overflow-y-auto rounded-xl border border-card-border">
            <JobNextRuns jobId={id} status={jobStatus?.status} />
          </div>
        </div>
        <JobRecentRuns jobId={id} />
        <Accordion type="single" collapsible>
          <AccordionItem value="advanced-settings">
            <AccordionTrigger className="-ml-2">
              <div className="hover:bg-gray-100 dark:hover:bg-gray-800 p-2 rounded-lg">
                Advanced Settings
              </div>
            </AccordionTrigger>
            <AccordionContent>
              <div className="flex flex-col gap-3">
                <div>
                  <WorkflowSettingsCard
                    job={data.job}
                    mutate={(newjob) => {
                      queryclient.setQueryData(
                        createConnectQueryKey(getJob, { id }),
                        new GetJobResponse({ job: newjob })
                      );
                    }}
                  />
                </div>
                <div>
                  <ActivitySyncOptionsCard
                    job={data.job}
                    mutate={(newjob) => {
                      queryclient.setQueryData(
                        createConnectQueryKey(getJob, { id }),
                        new GetJobResponse({ job: newjob })
                      );
                    }}
                  />
                </div>
              </div>
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>
    </div>
  );
}
