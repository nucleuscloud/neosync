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
import { Job, JobService } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { WorkflowSettingsSchema } from '../../../new/job/job-form-validations';
import { toWorkflowOptions } from '../../util';

interface Props {
  job: Job;
  mutate: (newjob: Job) => void;
}

export default function WorkflowSettingsCard({
  job,
  mutate,
}: Props): ReactElement<any> {
  const form = useForm<WorkflowSettingsSchema>({
    mode: 'onChange',
    resolver: yupResolver<WorkflowSettingsSchema>(WorkflowSettingsSchema),
    values: {
      runTimeout: job?.workflowOptions?.runTimeout
        ? convertNanosecondsToMinutes(job.workflowOptions.runTimeout)
        : 0,
    },
  });
  const { account } = useAccount();
  const { mutateAsync: updateJobWorkflowOptions } = useMutation(
    JobService.method.setJobWorkflowOptions
  );

  async function onSubmit(values: WorkflowSettingsSchema) {
    if (!account?.id) {
      return;
    }
    try {
      const resp = await updateJobWorkflowOptions({
        id: job.id,
        worfklowOptions: toWorkflowOptions(values),
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
          <CardContent>
            <FormField
              control={form.control}
              name="runTimeout"
              render={({ field }) => (
                <FormItem className="pt-4">
                  <FormLabel> Job Run Timeout</FormLabel>
                  <FormDescription>
                    The maximum length of time (in minutes) that a single job
                    run is allowed to span before it times out. <code>0</code>{' '}
                    means no overall timeout.
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
