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
import { convertNanosecondsToMinutes, getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { Job } from '@neosync/sdk';
import { setJobSyncOptions } from '@neosync/sdk/connectquery';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { ActivityOptionsFormValues } from '../../../new/job/job-form-validations';
import { toActivityOptions } from '../../util';

interface Props {
  job: Job;
  mutate: (newjob: Job) => void;
}

export default function ActivitySyncOptionsCard({
  job,
  mutate,
}: Props): ReactElement {
  const form = useForm<ActivityOptionsFormValues>({
    mode: 'onChange',
    resolver: yupResolver<ActivityOptionsFormValues>(ActivityOptionsFormValues),
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
  const { mutateAsync: updateJobSyncActivityOptions } =
    useMutation(setJobSyncOptions);

  async function onSubmit(values: ActivityOptionsFormValues) {
    if (!account?.id) {
      return;
    }
    try {
      const resp = await updateJobSyncActivityOptions({
        id: job.id,
        syncOptions: toActivityOptions(values),
      });
      toast.success('Successfully updated job workflow options!');
      if (resp.job) {
        mutate(resp.job);
      }
    } catch (err) {
      console.error(err);
      toast.error('Unable to update job workflow options', {
        description: getErrorMessage(err),
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
                      {`When exceeded, the retries stop even if they're not
                      expired yet. If not set or set to 0, it means unlimited
                      retry attemps and we rely on the max table timeout
                      including retries to know when to stop.`}
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
