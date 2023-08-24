'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import { useAccount } from '@/components/contexts/account-context';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import {
  CreateJobRequest,
  CreateJobResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { yupResolver } from '@hookform/resolvers/yup';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import {
  DefineFormValues,
  FLOW_FORM_SCHEMA,
  FlowFormValues,
  FormValues,
} from '../schema';

export default function Page({ searchParams }: PageProps): ReactElement {
  const router = useRouter();
  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = searchParams?.sessionId ?? '';
  const [defaultValues] = useSessionStorage<FlowFormValues>(
    `${sessionPrefix}-new-job-flow`,
    {
      sourceId: '',
      destinationId: '',
    }
  );

  const form = useForm({
    resolver: yupResolver<FlowFormValues>(FLOW_FORM_SCHEMA),
    defaultValues,
  });
  useFormPersist(`${sessionPrefix}-new-job-flow`, {
    watch: form.watch,
    setValue: form.setValue,
    storage: window.sessionStorage,
  });
  const account = useAccount();
  const { isLoading: isConnectionsLoading, data: connectionsData } =
    useGetConnections(account?.id ?? '');
  const [defineFormValues] = useSessionStorage<DefineFormValues>(
    `${sessionPrefix}-new-job-define`,
    { jobName: '' }
  );

  const connections = connectionsData?.connections ?? [];

  async function onSubmit(values: FlowFormValues) {
    if (!account?.id) {
      return;
    }
    try {
      const job = await createNewJob(account.id, {
        define: defineFormValues,
        flow: values,
        schema: {},
      });
      if (job.job?.id) {
        router.push(`/jobs/${job.job.id}`);
      } else {
        router.push(`/jobs`);
      }
    } catch (err) {
      console.error(err);
    }
    // router.push(`/new/job/schema?sessionId=${sessionPrefix}`);
  }

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
            name="sourceId"
            render={({ field }) => (
              <FormItem>
                <FormControl>
                  {/* <Input placeholder="Source ID" {...field} /> */}
                  {isConnectionsLoading ? (
                    <Skeleton />
                  ) : (
                    <Select onValueChange={field.onChange} value={field.value}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {connections
                          .filter(
                            (c) => c.id !== form.getValues().destinationId
                          )
                          .map((connection) => (
                            <SelectItem
                              className="cursor-pointer"
                              key={connection.id}
                              value={connection.id}
                            >
                              {connection.name}
                            </SelectItem>
                          ))}
                      </SelectContent>
                    </Select>
                  )}
                </FormControl>
                <FormDescription>
                  The location of the source data set.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="destinationId"
            render={({ field }) => (
              <FormItem>
                <FormControl>
                  {/* <Input placeholder="Source ID" {...field} /> */}
                  {isConnectionsLoading ? (
                    <Skeleton />
                  ) : (
                    <Select onValueChange={field.onChange} value={field.value}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {connections
                          .filter((c) => c.id !== form.getValues().sourceId)
                          .map((connection) => (
                            <SelectItem
                              className="cursor-pointer"
                              key={connection.id}
                              value={connection.id}
                            >
                              {connection.name}
                            </SelectItem>
                          ))}
                      </SelectContent>
                    </Select>
                  )}
                </FormControl>
                <FormDescription>
                  The location of the destination data set.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          {/* <FormField
            control={form.control}
            name="cronSchedule"
            render={({ field }) => (
              <FormItem>
                <FormControl>
                  <Input placeholder="Cron Schedule" {...field} />
                </FormControl>
                <FormDescription>
                  The schedule to run the job against if not a oneoff.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          /> */}

          <div className="flex flex-row gap-1 justify-between">
            <Button type="button" onClick={() => router.back()}>
              Back
            </Button>
            <Button type="submit">Next</Button>
          </div>
        </form>
      </Form>
    </OverviewContainer>
  );
}

async function createNewJob(
  accountId: string,
  formData: FormValues
): Promise<CreateJobResponse> {
  const res = await fetch(`/api/jobs`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new CreateJobRequest({
        accountId,
        jobName: formData.define.jobName,
        cronSchedule: formData.define.cronSchedule,
        haltOnNewColumnAddition: false,
        mappings: [],
        connectionSourceId: formData.flow.sourceId,
        connectionDestinationIds: [formData.flow.destinationId],
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CreateJobResponse.fromJson(await res.json());
}
