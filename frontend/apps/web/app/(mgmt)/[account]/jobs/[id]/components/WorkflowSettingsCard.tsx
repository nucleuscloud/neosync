'use client';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
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
import { useToast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  Job,
  SetJobWorkflowOptionsRequest,
  SetJobWorkflowOptionsResponse,
  WorkflowOptions,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';

const WORKFLOW_SETTINGS_FORM = Yup.object({
  runTimeout: Yup.number().optional().min(0),
});

type FormValues = Yup.InferType<typeof WORKFLOW_SETTINGS_FORM>;

interface Props {
  job: Job;
  mutate: (newjob: Job) => void;
}

export default function WorkflowSettingsCard({
  job,
  mutate,
}: Props): ReactElement {
  const { toast } = useToast();
  const form = useForm<FormValues>({
    mode: 'onChange',
    resolver: yupResolver<FormValues>(WORKFLOW_SETTINGS_FORM),
    values: {
      runTimeout: job?.workflowOptions?.runTimeout
        ? convertNanosecondsToMinutes(job.workflowOptions.runTimeout)
        : 0,
    },
  });
  const { account } = useAccount();

  async function onSubmit(values: FormValues) {
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
      <CardHeader>
        <CardTitle>Workflow Options</CardTitle>
        <CardDescription>
          Advanced workflow settings for configuring run timeouts and other
          settings in the future.
        </CardDescription>
      </CardHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)}>
          <CardContent>
            <FormField
              control={form.control}
              name="runTimeout"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Run Timeout</FormLabel>
                  <FormDescription>
                    The maximum length of time in minutes that a single job run
                    is allowed to span before it times out.
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
  values: FormValues
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
