'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { yupResolver } from '@hookform/resolvers/yup';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useState, use } from 'react';
import { useForm } from 'react-hook-form';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps, { getJobProgressSteps } from '../JobsProgressSteps';
import {
  DefineFormValues,
  DefineFormValuesContext,
  NewJobType,
} from '../job-form-validations';

import FormPersist from '@/app/(mgmt)/FormPersist';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Separator } from '@/components/ui/separator';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { getSingleOrUndefined } from '@/libs/utils';
import { useMutation } from '@connectrpc/connect-query';
import { JobService } from '@neosync/sdk';
import { usePostHog } from 'posthog-js/react';
import { DEFAULT_CRON_STRING } from '../../../jobs/[id]/components/ScheduleCard';
import { getNewJobSessionKeys } from '../../../jobs/util';

export default function Page(props: PageProps): ReactElement<any> {
  const searchParams = use(props.searchParams);
  const router = useRouter();
  const { account } = useAccount();
  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/${account?.name}/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = getSingleOrUndefined(searchParams?.sessionId) ?? '';
  const sessionKeys = getNewJobSessionKeys(sessionPrefix);

  const formKey = sessionKeys.global.define;
  const [defaultValues] = useSessionStorage<DefineFormValues>(formKey, {
    jobName: '',
    cronSchedule: '',
    initiateJobRun: false,
    workflowSettings: {},
    syncActivityOptions: {
      startToCloseTimeout: 10,
      retryPolicy: {
        maximumAttempts: 1,
      },
    },
  });
  const [isScheduleEnabled, setIsScheduleEnabled] = useState<boolean>(false);

  // run once to check if schedule should be enabled. (only want to run this a single time)
  useEffect(() => {
    if (defaultValues.cronSchedule && !isScheduleEnabled) {
      setIsScheduleEnabled(true);
    }
  }, []);

  const { mutateAsync: isJobNameAvailableAsync } = useMutation(
    JobService.method.isJobNameAvailable
  );
  const form = useForm<DefineFormValues, DefineFormValuesContext>({
    mode: 'onChange',
    resolver: yupResolver<DefineFormValues>(DefineFormValues),
    defaultValues,
    context: {
      accountId: account?.id ?? '',
      isJobNameAvailable: isJobNameAvailableAsync,
    },
  });

  const newJobType = getNewJobType(getSingleOrUndefined(searchParams?.jobType));
  const posthog = usePostHog();

  async function onSubmit(_values: DefineFormValues) {
    if (!isScheduleEnabled) {
      form.setValue('cronSchedule', '');
    }
    posthog.capture('New Job Flow Define Complete', { jobType: newJobType });
    if (newJobType === 'generate-table') {
      router.push(
        `/${account?.name}/new/job/generate/single/connect?sessionId=${sessionPrefix}`
      );
    } else if (newJobType === 'ai-generate-table') {
      router.push(
        `/${account?.name}/new/job/aigenerate/single/connect?sessionId=${sessionPrefix}`
      );
    } else {
      router.push(
        `/${account?.name}/new/job/connect?sessionId=${sessionPrefix}`
      );
    }
  }

  return (
    <div
      id="newjobdefine"
      className="px-12 md:px-24 lg:px-48 xl:px-64 flex flex-col gap-5"
    >
      <FormPersist formKey={formKey} form={form} />
      <OverviewContainer
        Header={
          <PageHeader
            header="Define"
            progressSteps={
              <JobsProgressSteps
                steps={getJobProgressSteps(newJobType)}
                stepName={'define'}
              />
            }
          />
        }
        containerClassName="connect-page"
      >
        <div />
      </OverviewContainer>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)}>
          <FormField
            control={form.control}
            name="jobName"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormDescription>The unique name of the job.</FormDescription>
                <FormControl>
                  <Input placeholder="prod-to-stage" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="cronSchedule"
            render={({ field }) => (
              <FormItem>
                <div className="flex flex-row items-center gap-2">
                  <FormLabel>Schedule</FormLabel>
                  <Switch
                    checked={isScheduleEnabled}
                    onCheckedChange={(isChecked) => {
                      setIsScheduleEnabled(isChecked);
                      if (!isChecked) {
                        form.resetField('cronSchedule', {
                          keepError: false,
                        });
                      }
                    }}
                  />
                </div>
                <FormDescription>
                  Define a cron schedule to run this job. If disabled, the job
                  will be paused and a default cron will be set
                </FormDescription>
                <FormControl>
                  <Input
                    placeholder={DEFAULT_CRON_STRING}
                    {...field}
                    disabled={!isScheduleEnabled}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <div>
            <FormField
              name="initiateJobRun"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Initiate Job Run</FormLabel>
                  <FormDescription>
                    Initiates a single job run immediately after the job is
                    created.
                  </FormDescription>
                  <FormControl>
                    <ToggleGroup
                      type="single"
                      className="flex justify-start"
                      onValueChange={(value) =>
                        field.onChange(value === 'true')
                      }
                      value={field.value ? 'true' : 'false'}
                    >
                      <ToggleGroupItem value="false">No</ToggleGroupItem>
                      <ToggleGroupItem value="true">Yes</ToggleGroupItem>
                    </ToggleGroup>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
          <Accordion type="single" collapsible className="w-full">
            <AccordionItem value="settings">
              <AccordionTrigger className="-ml-2">
                <div className="hover:bg-gray-100 dark:hover:bg-gray-800 p-2 rounded-lg">
                  Advanced Settings
                </div>
              </AccordionTrigger>
              <AccordionContent>
                <Separator />
                <div className="flex flex-col pt-6 p-2">
                  <FormField
                    control={form.control}
                    name="workflowSettings.runTimeout"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel> Job Run Timeout</FormLabel>
                        <FormDescription>
                          The maximum length of time (in minutes) that a single
                          job run is allowed to span before it times out.{' '}
                          <code>0</code> means no overall timeout.
                        </FormDescription>
                        <FormControl>
                          <Input
                            type="number"
                            {...field}
                            value={field.value || 0}
                            onChange={(e) => {
                              field.onChange(e.target.valueAsNumber);
                            }}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <div className="flex flex-col">
                    <FormField
                      control={form.control}
                      name="syncActivityOptions.startToCloseTimeout"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Table Sync Timeout</FormLabel>
                          <FormDescription>
                            The maximum amount of time (in minutes) a single
                            table synchronization may run before it times out.
                            This may need tuning depending on your datasize, and
                            should be able to contain the table that contains
                            the largest amount of data. This timeout is applied
                            per retry.
                          </FormDescription>
                          <FormControl>
                            <Input
                              type="number"
                              {...field}
                              value={field.value || 0}
                              onChange={(e) => {
                                field.onChange(e.target.valueAsNumber);
                              }}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="syncActivityOptions.scheduleToCloseTimeout"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>
                            Max Table Timeout including retries
                          </FormLabel>
                          <FormDescription>
                            The total time (in minutes) that a single table sync
                            is allowed to run,{' '}
                            <strong>
                              <u>including</u>
                            </strong>{' '}
                            retries. 0 means no timeout.
                          </FormDescription>
                          <FormControl>
                            <Input
                              type="number"
                              {...field}
                              value={field.value || 0}
                              onChange={(e) => {
                                field.onChange(e.target.valueAsNumber);
                              }}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                  <div>
                    <FormField
                      control={form.control}
                      name="syncActivityOptions.retryPolicy.maximumAttempts"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Maximum Retry Attempts</FormLabel>
                          <FormDescription>
                            {`When exceeded, the retries stop even if they're not
                            expired yet. If not set or set to 0, it means
                            unlimited retry attemps and we rely on the max table
                            timeout including retries to know when to stop.`}
                          </FormDescription>
                          <FormControl>
                            <Input
                              type="number"
                              {...field}
                              value={field.value || 0}
                              onChange={(e) => {
                                field.onChange(e.target.valueAsNumber);
                              }}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                </div>
              </AccordionContent>
            </AccordionItem>
          </Accordion>
          <div className="flex flex-row justify-between pt-10 ">
            <Button
              variant="outline"
              type="reset"
              onClick={() => router.push(`/${account?.name}/new/job`)}
            >
              Back
            </Button>
            <Button type="submit" disabled={!form.formState.isValid}>
              Next
            </Button>
          </div>
        </form>
      </Form>
    </div>
  );
}

function getNewJobType(jobtype?: string): NewJobType {
  switch (jobtype) {
    case 'generate-table':
    case 'ai-generate-table':
    case 'data-sync':
      return jobtype as NewJobType;
    default:
      return 'data-sync';
  }
}
