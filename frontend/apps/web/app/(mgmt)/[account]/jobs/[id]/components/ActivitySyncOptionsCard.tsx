'use client';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter } from '@/components/ui/card';
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
import { useToast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  ActivityOptions,
  Job,
  RetryPolicy,
  SetJobSyncOptionsRequest,
  SetJobWorkflowOptionsResponse,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { ActivityOptionsSchema } from '../../../new/job/schema';

interface Props {
  job: Job;
  mutate: (newjob: Job) => void;
}

export default function ActivitySyncOptionsCard({
  job,
  mutate,
}: Props): ReactElement {
  const { toast } = useToast();
  const form = useForm<ActivityOptionsSchema>({
    mode: 'onChange',
    resolver: yupResolver<ActivityOptionsSchema>(ActivityOptionsSchema),
    values: {
      scheduleToCloseTimeout: job?.syncOptions?.scheduleToCloseTimeout
        ? convertNanosecondsToMinutes(job.syncOptions.scheduleToCloseTimeout)
        : 0,
      startToCloseTimeout: job?.syncOptions?.startToCloseTimeout
        ? convertNanosecondsToMinutes(job.syncOptions.startToCloseTimeout)
        : 10,
      retryPolicy: {
        maximumAttempts: job?.syncOptions?.retryPolicy?.maximumAttempts ?? 1,
      },
    },
  });
  const { account } = useAccount();

  async function onSubmit(values: ActivityOptionsSchema) {
    if (!account?.id) {
      return;
    }
    try {
      const resp = await updateJobSyncActivityOptions(
        account.id,
        job.id,
        values
      );
      toast({
        title: 'Successfully updated job workflow options!',
        variant: 'success',
      });
      if (resp.job) {
        mutate(resp.job);
      }
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to update job workflow options',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <Card className="overflow-hidden">
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)}>
          <CardContent className="space-y-6">
            <div className="flex flex-col gap-6 ">
              <FormField
                control={form.control}
                name="startToCloseTimeout"
                render={({ field }) => (
                  <FormItem>
                    <div className="pt-4">
                      <FormLabel>Table Sync Timeout</FormLabel>
                      <FormDescription>
                        The maximum amount of time (in minutes) a single table
                        synchronization may run before it times out. This may
                        need tuning depending on your datasize, and should be
                        able to contain the table that contains the largest
                        amount of data. This timeout is applied per retry.
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
                name="scheduleToCloseTimeout"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Max Table Timeout including retries</FormLabel>
                    <FormDescription>
                      The total time (in minutes) that a single table sync is
                      allowed to run,{' '}
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
                name="retryPolicy.maximumAttempts"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Maximum Retry Attempts</FormLabel>
                    <FormDescription>
                      When exceeded, the retries stop even if they're not
                      expired yet. If not set or set to 0, it means unlimited
                      retry attemps and we rely on the max table timeout
                      including retries to know when to stop.
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
          <CardFooter className="bg-muted flex py-2 justify-center">
            <div className="flex flex-row items-center justify-end w-full">
              <Button type="submit" disabled={!form.formState.isValid}>
                Save
              </Button>
            </div>
          </CardFooter>
        </form>
      </Form>
    </Card>
  );
}

async function updateJobSyncActivityOptions(
  accountId: string,
  jobId: string,
  values: ActivityOptionsSchema
): Promise<SetJobWorkflowOptionsResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${jobId}/syncoptions`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(
        new SetJobSyncOptionsRequest({
          id: jobId,
          syncOptions: new ActivityOptions({
            startToCloseTimeout:
              values.startToCloseTimeout !== undefined &&
              values.startToCloseTimeout > 0
                ? convertMinutesToNanoseconds(values.startToCloseTimeout)
                : undefined,
            scheduleToCloseTimeout:
              values.scheduleToCloseTimeout !== undefined &&
              values.scheduleToCloseTimeout > 0
                ? convertMinutesToNanoseconds(values.scheduleToCloseTimeout)
                : undefined,
            retryPolicy: new RetryPolicy({
              maximumAttempts: values.retryPolicy?.maximumAttempts,
            }),
          }),
        })
      ),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return SetJobWorkflowOptionsResponse.fromJson(await res.json());
}

const NANOS_PER_SECOND = BigInt(1000000000);
const SECONDS_PER_MIN = BigInt(60);

// if the duration is too large to convert to minutes, it will return the max safe integer
function convertNanosecondsToMinutes(duration: bigint): number {
  // Convert nanoseconds to minutes
  const minutesBigInt = duration / NANOS_PER_SECOND / SECONDS_PER_MIN;

  // Check if the result is within the safe range for JavaScript numbers
  if (minutesBigInt <= BigInt(Number.MAX_SAFE_INTEGER)) {
    return Number(minutesBigInt);
  } else {
    // Handle the case where the number of minutes is too large
    console.warn(
      'The number of minutes is too large for a safe JavaScript number. Returning as BigInt.'
    );
    return Number.MAX_SAFE_INTEGER;
  }
}

// Convert minutes to BigInt to ensure precision in multiplication
function convertMinutesToNanoseconds(minutes: number): bigint {
  const minutesBigInt = BigInt(minutes);
  return minutesBigInt * SECONDS_PER_MIN * NANOS_PER_SECOND;
}
