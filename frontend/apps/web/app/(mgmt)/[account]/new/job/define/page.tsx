'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
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
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps, {
  DATA_GEN_STEPS,
  DATA_SYNC_STEPS,
} from '../JobsProgressSteps';
import { NewJobType } from '../page';
import { DEFINE_FORM_SCHEMA, DefineFormValues } from '../schema';

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Label } from '@/components/ui/label';

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
  const [disableSchedule, setDisableSchedule] = useState<boolean>(false);

  const form = useForm<DefineFormValues>({
    mode: 'onChange',
    resolver: yupResolver<DefineFormValues>(DEFINE_FORM_SCHEMA),
    defaultValues,
    context: { accountId: account?.id ?? '', showSchedule: disableSchedule },
  });

  useFormPersist(`${sessionPrefix}-new-job-define`, {
    watch: form.watch,
    setValue: form.setValue,
    storage: isBrowser() ? window.sessionStorage : undefined,
  });

  const newJobType = getNewJobType(getSingleOrUndefined(searchParams?.jobType));

  async function onSubmit(_values: DefineFormValues) {
    if (!disableSchedule) {
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

  useEffect(() => {
    if (form.getValues('cronSchedule')) {
      setDisableSchedule(true);
    }
  }, [disableSchedule, form]);

  return (
    <div
      id="newjobdefine"
      className="px-12 md:px-24 lg:px-32 flex flex-col gap-5"
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
          <FormField
            control={form.control}
            name="cronSchedule"
            render={({ field }) => (
              <FormItem>
                <div className="flex flex-row items-center gap-2">
                  <FormLabel>Schedule</FormLabel>
                  <Switch
                    checked={disableSchedule}
                    onCheckedChange={() => {
                      disableSchedule
                        ? setDisableSchedule(false)
                        : setDisableSchedule(true);
                    }}
                  />
                </div>
                <FormDescription>
                  Define a cron schedule to run this job. If disabled, the job
                  will need to be manually executed.
                </FormDescription>
                <FormControl>
                  <Input
                    placeholder="0 0 * * * "
                    {...field}
                    disabled={!disableSchedule}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <div>
            <div>
              <FormField
                name="initiateJobRun"
                render={({ field }) => (
                  <FormItem>
                    <FormControl>
                      <div className="flex flex-row items-center justify-between">
                        <div className="space-y-0.5">
                          <Label className="text-sm">Initiate Job Run</Label>
                          <p className="text-xs text-muted-foreground">
                            Initiates a single job run immediately after job is
                            created.
                          </p>
                        </div>
                        <Switch
                          checked={field.value || false}
                          onCheckedChange={field.onChange}
                        />
                      </div>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </div>

          <Accordion type="single" collapsible className="w-full">
            <AccordionItem value="settings">
              <AccordionTrigger>
                <Button
                  variant="ghost"
                  className="hover:bg-gray-100 -ml-4"
                  type="button"
                >
                  Advanced Settings
                </Button>
              </AccordionTrigger>
              <AccordionContent>
                <div className="pt-6">
                  <Card className="overflow-hidden">
                    <CardContent className="pt-4">
                      <FormField
                        control={form.control}
                        name="workflowSettings.runTimeout"
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel> Job Run Timeout</FormLabel>
                            <FormDescription>
                              The maximum length of time, in minutes, that a
                              single job run is allowed to span before it times
                              out. 0 means no overall timeout.
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
                    </CardContent>
                  </Card>
                </div>
                <div className="pt-10">
                  <Card className="overflow-hidden">
                    <CardHeader>
                      <CardTitle>Table Synchronization Options</CardTitle>
                      <CardDescription>
                        Advanced settings that are applied to every table
                        synchronization. A table sync timout or max timeout must
                        be configured to run.
                      </CardDescription>
                    </CardHeader>

                    <CardContent className="space-y-6">
                      <div className="flex flex-col md:flex-row gap-3 justify-between">
                        <FormField
                          control={form.control}
                          name="syncActivityOptions.startToCloseTimeout"
                          render={({ field }) => (
                            <FormItem className="w-full md:w-1/2 flex flex-col gap-2 justify-between space-y-0">
                              <div>
                                <FormLabel>Table Sync Timeout</FormLabel>
                                <FormDescription>
                                  The max amount of time that a single table
                                  synchronization may run before it times out.
                                  This may need tuning depending on your
                                  datasize, and should be able to contain the
                                  table that contains the largest amount of
                                  data. This timeout is applied per retry.
                                </FormDescription>
                              </div>
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
                            <FormItem className="w-full md:w-1/2 flex flex-col gap-2 justify-between space-y-0">
                              <div>
                                <FormLabel>
                                  Max Table Timeout including retries
                                </FormLabel>
                                <FormDescription>
                                  Total time in minutes that a single table sync
                                  is allowed to run, including retires. 0 means
                                  no timeout.
                                </FormDescription>
                              </div>
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
                        <h2 className="text-xl font-semibold tracking-tight">
                          Retry Policy
                        </h2>
                        <FormField
                          control={form.control}
                          name="syncActivityOptions.retryPolicy.maximumAttempts"
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>Maximum Attempts</FormLabel>
                              <FormDescription>
                                Maximum number of attempts. When exceeded the
                                retries stop even if not expired yet. If not set
                                or set to 0, it means unlimited, and relies on
                                activity the max table timeout including retries
                                to know when to stop.
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
                    </CardContent>
                  </Card>
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
