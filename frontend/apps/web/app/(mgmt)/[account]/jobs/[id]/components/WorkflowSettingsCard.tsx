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
import {
  convertMinutesToNanoseconds,
  convertNanosecondsToMinutes,
  getErrorMessage,
} from '@/util/util';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  Job,
  SetJobWorkflowOptionsRequest,
  SetJobWorkflowOptionsResponse,
  WorkflowOptions,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { WorkflowSettingsSchema } from '../../../new/job/schema';

interface Props {
  job: Job;
  mutate: (newjob: Job) => void;
}

export default function WorkflowSettingsCard({
  job,
  mutate,
}: Props): ReactElement {
  const { toast } = useToast();
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

  async function onSubmit(values: WorkflowSettingsSchema) {
    if (!account?.id) {
      return;
    }
    try {
      const resp = await updateJobWorkflowOptions(account.id, job.id, values);
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

async function updateJobWorkflowOptions(
  accountId: string,
  jobId: string,
  values: WorkflowSettingsSchema
): Promise<SetJobWorkflowOptionsResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${jobId}/workflowoptions`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(
        new SetJobWorkflowOptionsRequest({
          id: jobId,
          worfklowOptions: new WorkflowOptions({
            runTimeout:
              values.runTimeout !== undefined && values.runTimeout > 0
                ? convertMinutesToNanoseconds(values.runTimeout)
                : undefined,
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
