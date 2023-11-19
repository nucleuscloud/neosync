'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
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
import { yupResolver } from '@hookform/resolvers/yup';
import NeoCron from 'neocron';
import 'neocron/dist/src/globals.css';
import { usePathname, useRouter } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps from '../JobsProgressSteps';
import { DEFINE_FORM_SCHEMA, DefineFormValues } from '../schema';

export default function Page({ searchParams }: PageProps): ReactElement {
  const router = useRouter();
  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = searchParams?.sessionId ?? '';
  const [defaultValues] = useSessionStorage<DefineFormValues>(
    `${sessionPrefix}-new-job-define`,
    {
      jobName: '',
      cronSchedule: '* * * * *',
      initiateJobRun: false,
    }
  );

  const form = useForm({
    resolver: yupResolver<DefineFormValues>(DEFINE_FORM_SCHEMA),
    defaultValues,
  });

  const isBrowser = () => typeof window !== 'undefined';

  useFormPersist(`${sessionPrefix}-new-job-define`, {
    watch: form.watch,
    setValue: form.setValue,
    storage: isBrowser() ? window.sessionStorage : undefined,
  });

  async function onSubmit(_values: DefineFormValues) {
    router.push(`/new/job/connect?sessionId=${sessionPrefix}`);
  }

  const [isClient, setIsClient] = useState(false);
  useEffect(() => {
    // This code runs after mount, indicating we're on the client
    setIsClient(true);
  }, []);

  //check if there is somethign in the values for this page and if so then set this to complete

  const params = usePathname();
  const [stepName, _] = useState<string>(params.split('/').pop() ?? '');

  const [isCompleted, setIsCompleted] = useState<boolean>(
    form.getValues('jobName') !== '' ? true : false
  );

  return (
    <div id="newjobdefine" className="px-12 md:px-24 lg:px-32">
      <OverviewContainer
        Header={
          <PageHeader
            header="Create a new Job"
            description="Define a new job to move, transform, or scan data"
          />
        }
      >
        <JobsProgressSteps stepName={stepName} />
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
            {isClient && (
              <Controller
                control={form.control}
                name="cronSchedule"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Schedule</FormLabel>
                    <FormDescription>
                      Define a schedule to run this job.
                    </FormDescription>
                    <FormControl>
                      <NeoCron
                        cronString={field.value ?? ''}
                        defaultCronString="* * * * *"
                        setCronString={field.onChange}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
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
            <div className="flex flex-row justify-end">
              <Button type="submit">Next</Button>
            </div>
          </form>
        </Form>
      </OverviewContainer>
    </div>
  );
}
