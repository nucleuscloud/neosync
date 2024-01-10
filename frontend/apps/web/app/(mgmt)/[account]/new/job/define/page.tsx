'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SwitchCard from '@/components/switches/SwitchCard';
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
    }
  );
  const [showSchedule, setShowSchedule] = useState<boolean>(false);

  const form = useForm<DefineFormValues>({
    mode: 'onChange',
    resolver: yupResolver<DefineFormValues>(DEFINE_FORM_SCHEMA),
    defaultValues,
    context: { accountId: account?.id ?? '', showSchedule: showSchedule },
  });

  const isBrowser = () => typeof window !== 'undefined';

  useFormPersist(`${sessionPrefix}-new-job-define`, {
    watch: form.watch,
    setValue: form.setValue,
    storage: isBrowser() ? window.sessionStorage : undefined,
  });

  const newJobType = getNewJobType(getSingleOrUndefined(searchParams?.jobType));

  async function onSubmit(_values: DefineFormValues) {
    if (!showSchedule) {
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
      setShowSchedule(true);
    }
  }, [setShowSchedule, form]);

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
                  <Input placeholder="Job Name" {...field} />
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
                    checked={showSchedule}
                    onCheckedChange={() => {
                      showSchedule
                        ? setShowSchedule(false)
                        : setShowSchedule(true);
                    }}
                  />
                </div>
                <FormDescription>
                  Define a cron schedule to run this job. If disabled, the job
                  will need to be manually triggered.
                </FormDescription>
                <FormControl>
                  {showSchedule && (
                    <Input placeholder="0 0 * * * " {...field} />
                  )}
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <div>
            <FormLabel>Settings</FormLabel>
            <div className="pt-4">
              <FormField
                name="initiateJobRun"
                render={({ field }) => (
                  <FormItem>
                    <FormControl>
                      <SwitchCard
                        isChecked={field.value || false}
                        onCheckedChange={field.onChange}
                        title="Initiate Job Run"
                        description="Initiates a single job run immediately after job is created."
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </div>
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
