'use client';
import { PageProps } from '@/components/types';

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { create } from '@bufbuild/protobuf';
import { createConnectQueryKey, useQuery } from '@connectrpc/connect-query';
import { GetJobResponseSchema, JobService } from '@neosync/sdk';
import { useQueryClient } from '@tanstack/react-query';
import { ReactElement, use } from 'react';
import ActivitySyncOptionsCard from './components/ActivitySyncOptionsCard';
import JobNextRuns from './components/NextRuns';
import JobRecentRuns from './components/RecentRuns';
import JobScheduleCard from './components/ScheduleCard';
import WorkflowSettingsCard from './components/WorkflowSettingsCard';
import JobIdSkeletonForm from './JobIdSkeletonForm';

export default function Page(props: PageProps): ReactElement<any> {
  const params = use(props.params);
  const id = params?.id ?? '';
  const { data, isLoading } = useQuery(
    JobService.method.getJob,
    { id },
    { enabled: !!id }
  );
  const { data: jobStatus } = useQuery(
    JobService.method.getJobStatus,
    { jobId: id },
    { enabled: !!id }
  );
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
          <div className="grow basis-3/4">
            <JobScheduleCard
              job={data.job}
              mutate={(newjob) => {
                queryclient.setQueryData(
                  createConnectQueryKey({
                    schema: JobService.method.getJob,
                    input: { id },
                    cardinality: undefined,
                  }),
                  create(GetJobResponseSchema, { job: newjob })
                );
              }}
            />
          </div>
          <div className="grow basis-1/4 overflow-y-auto rounded-xl border border-card-border">
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
                        createConnectQueryKey({
                          schema: JobService.method.getJob,
                          input: { id },
                          cardinality: undefined,
                        }),
                        create(GetJobResponseSchema, { job: newjob })
                      );
                    }}
                  />
                </div>
                <div>
                  <ActivitySyncOptionsCard
                    job={data.job}
                    mutate={(newjob) => {
                      queryclient.setQueryData(
                        createConnectQueryKey({
                          schema: JobService.method.getJob,
                          input: { id },
                          cardinality: undefined,
                        }),
                        create(GetJobResponseSchema, { job: newjob })
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
