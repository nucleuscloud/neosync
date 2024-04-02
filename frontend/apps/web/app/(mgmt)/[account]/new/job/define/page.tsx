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
import { yupResolver } from '@hookform/resolvers/yup';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps, {
  DATA_GEN_STEPS,
  DATA_SYNC_STEPS,
} from '../JobsProgressSteps';
import { NewJobType } from '../page';
import { DEFINE_FORM_SCHEMA, DefineFormValues } from '../schema';

import FormError from '@/components/FormError';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { DEFAULT_CRON_STRING } from '../../../jobs/[id]/components/ScheduleCard';

const isBrowser = () => typeof window !== 'undefined';

export default function Page({ searchParams }: PageProps): ReactElement {
  const router = useRouter();
  const { account } = useAccount();
  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/${account?.name}/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = searchParams?.sessionId ?? '';
  const [defaultValues] = useSessionStorage<DefineFormValues>(
    `${sessionPrefix}-new-job-define`,
    {
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
    }
  );
  const [isScheduleEnabled, setIsScheduleEnabled] = useState<boolean>(false);

  const form = useForm<DefineFormValues>({
    mode: 'onChange',
    resolver: yupResolver<DefineFormValues>(DEFINE_FORM_SCHEMA),
    defaultValues,
    context: {
      accountId: account?.id ?? '',
    },
  });

  useFormPersist(`${sessionPrefix}-new-job-define`, {
    watch: form.watch,
    setValue: form.setValue,
    storage: isBrowser() ? window.sessionStorage : undefined,
  });

  const newJobType = getNewJobType(getSingleOrUndefined(searchParams?.jobType));

  async function onSubmit(_values: DefineFormValues) {
    if (!isScheduleEnabled) {
      form.setValue('cronSchedule', '');
    }
    if (newJobType === 'generate-table') {
      router.push(
        `/${account?.name}/new/job/generate/single/connect?sessionId=${sessionPrefix}`
      );
    } else {
      router.push(
        `/${account?.name}/new/job/connect?sessionId=${sessionPrefix}`
      );
    }
  }

  const scheduleOptions = [
    {
      name: 'Daily',
      cron: '0 8 * * *',
    },
    {
      name: 'Weekly',
      cron: '0 8 * * 1',
    },
    {
      name: 'Monthly',
      cron: '0 8 1 * *',
    },
    {
      name: 'Custom',
      cron: 'custom',
    },
  ];

  const [customCron, setCustomCron] = useState<boolean>(false);
  const [cronScheduleName, setCronScheduleName] = useState<string>('');

  const handleSettingCronSchedule = (val: string) => {
    const cs = scheduleOptions.find((item) => item.name == val)?.cron;
    if (val == 'Custom') {
      setCustomCron(true);
      setCronScheduleName(val);
      form.setValue('cronSchedule', cs);
    } else {
      setCustomCron(false);
      form.setValue('cronSchedule', cs);
      form.clearErrors();
    }
  };

  // used to set the select if the user navigates bakc to the form
  useEffect(() => {
    const cron = form.getValues('cronSchedule');
    if (cron) {
      setIsScheduleEnabled(true);
    }
  }, [form.watch('cronSchedule')]);

  // used to handle setting the error states on the custom cron input
  useEffect(() => {
    if (form.formState.errors.cronSchedule) {
      form.setError('cronSchedule', {
        message:
          form.formState.errors.cronSchedule?.message ?? 'Invalid cron string',
      });
    }
  }, [form.formState.errors.cronSchedule?.message]);

  console.log('form', form.getValues('cronSchedule'));

  // TODO: update the value that gets set when you enabled and disable the schedule

  return (
    <div
      id="newjobdefine"
      className="px-12 md:px-24 lg:px-48 xl:px-64 flex flex-col gap-5"
    >
      <OverviewContainer
        Header={
          <PageHeader
            header="Define"
            progressSteps={
              <JobsProgressSteps
                steps={
                  newJobType === 'data-sync' ? DATA_SYNC_STEPS : DATA_GEN_STEPS
                }
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
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
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
          <Controller
            control={form.control}
            name="cronSchedule"
            render={({ field: { onChange, ...field } }) => (
              <FormItem>
                <div className="flex flex-row items-center gap-2">
                  <FormLabel>Schedule</FormLabel>
                  <Switch
                    checked={isScheduleEnabled}
                    onCheckedChange={(isChecked) => {
                      if (isChecked) {
                        setIsScheduleEnabled(false);
                        form.resetField('cronSchedule', {
                          keepError: false,
                          defaultValue: DEFAULT_CRON_STRING,
                        });
                        // setCronScheduleName('');
                      } else {
                        setIsScheduleEnabled(true);
                      }
                      setIsScheduleEnabled(isChecked);
                      // if (!isChecked) {
                      //   setIsScheduleEnabled(false);
                      //   form.resetField('cronSchedule', {
                      //     keepError: false,
                      //     defaultValue: DEFAULT_CRON_STRING,
                      //   });
                      // }
                    }}
                  />
                </div>
                <FormDescription>
                  Define a cron schedule to run this job. If disabled, the job
                  will be paused and a default schedule will be set.
                </FormDescription>
                <FormControl>
                  <div className="flex flex-col md:flex-row items-center gap-2">
                    <Select
                      onValueChange={(value) => {
                        handleSettingCronSchedule(value);
                      }}
                      value={
                        scheduleOptions.find(
                          (item) => item.cron == form.getValues('cronSchedule')
                        )?.name
                          ? scheduleOptions.find(
                              (item) =>
                                item.cron == form.getValues('cronSchedule')
                            )?.name
                          : customCron
                            ? 'Custom'
                            : ''
                      }
                    >
                      <SelectTrigger
                        disabled={!isScheduleEnabled}
                        className="min-w-[400px] w-[400px]"
                      >
                        <SelectValue placeholder="Set a schedule" />
                      </SelectTrigger>
                      <SelectContent>
                        {scheduleOptions.map((opts) => (
                          <SelectItem
                            className="cursor-pointer"
                            key={opts.name}
                            value={opts.name}
                          >
                            {opts.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    {customCron && (
                      <Input
                        placeholder={DEFAULT_CRON_STRING}
                        {...field}
                        // onChange={(value) => handleCronInput(value)}
                        // onChange={(value) => field.onChange(value)}
                        onChange={async ({ target: { value } }) => {
                          onChange(value);
                          await form.trigger('cronSchedule');
                        }}
                        disabled={!isScheduleEnabled}
                      />
                    )}
                  </div>
                  {/* {form.getValues('cronSchedule') && (
                    <div className="text-xs">
                      This schedule will run at 08:00{' '}
                      {getScheduleFromCronString(
                        form.getValues('cronSchedule') ?? ''
                      )}
                    </div>
                  )} */}
                </FormControl>
                <FormError
                  errorMessage={
                    form.formState.errors.cronSchedule?.message ?? ''
                  }
                />
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
                <div className="flex flex-col gap-6 pt-6">
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
                  <div className="flex flex-col gap-6">
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
                            When exceeded, the retries stop even if they're not
                            expired yet. If not set or set to 0, it means
                            unlimited retry attemps and we rely on the max table
                            timeout including retries to know when to stop.
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
          <div className="flex flex-row justify-between">
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
  return jobtype === 'generate-table' ? 'generate-table' : 'data-sync';
}
function getSingleOrUndefined(
  item: string | string[] | undefined
): string | undefined {
  if (!item) {
    return undefined;
  }
  const newItem = Array.isArray(item) ? item[0] : item;
  return !newItem || newItem === 'undefined' ? undefined : newItem;
}
