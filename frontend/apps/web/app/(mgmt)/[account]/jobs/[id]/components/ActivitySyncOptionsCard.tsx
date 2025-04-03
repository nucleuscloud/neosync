'use client';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter } from '@/components/ui/card';
import { Form } from '@/components/ui/form';
import { convertNanosecondsToMinutes, getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { Job, JobService } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import SyncActivityOptionsForm from '../../../new/job/define/components/WorkflowSettings';
import {
  ActivityOptionsFormValues,
  NewJobType,
} from '../../../new/job/job-form-validations';
import { toActivityOptions } from '../../util';
import { isPiiDetectJob } from '../util';

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
  const { mutateAsync: updateJobSyncActivityOptions } = useMutation(
    JobService.method.setJobSyncOptions
  );

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
          <CardContent className="space-y-6 mt-4">
            <SyncActivityOptionsForm
              value={form.watch() ?? {}}
              setValue={(value) => {
                if (value.retryPolicy != undefined) {
                  form.setValue(
                    'retryPolicy.maximumAttempts',
                    value.retryPolicy?.maximumAttempts
                  );
                }
                if (value.scheduleToCloseTimeout != undefined) {
                  form.setValue(
                    'scheduleToCloseTimeout',
                    value.scheduleToCloseTimeout
                  );
                }
                if (value.startToCloseTimeout != undefined) {
                  form.setValue(
                    'startToCloseTimeout',
                    value.startToCloseTimeout
                  );
                }
              }}
              errors={form.formState.errors}
              jobtype={getJobType(job)}
            />
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

function getJobType(job?: Job): NewJobType {
  if (job && isPiiDetectJob(job)) {
    return 'pii-detection';
  }
  return 'data-sync';
}
