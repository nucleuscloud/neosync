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
import { default as NeoCron } from 'neocron';
import 'neocron/dist/src/globals.css';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
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
    router.push(`/new/job/flow?sessionId=${sessionPrefix}`);
  }

  const [isClient, setIsClient] = useState(false);

  useEffect(() => {
    // This code runs after mount, indicating we're on the client
    setIsClient(true);
  }, []);

  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Create a new Job"
          description="Define a new job to move, transform, or scan data"
        />
      }
    >
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <FormField
            control={form.control}
            name="jobName"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl>
                  <Input placeholder="Job Name" {...field} />
                </FormControl>
                <FormDescription>The unique name of the job.</FormDescription>
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
          <div className="max-w-[500px]">
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
          <div className="flex flex-row justify-end">
            <Button type="submit">Next</Button>
          </div>
        </form>
      </Form>
    </OverviewContainer>
  );
}
