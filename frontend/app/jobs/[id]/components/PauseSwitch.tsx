'use client';
import SwitchCard from '@/components/switches/SwitchCard';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import { Textarea } from '@/components/ui/textarea';
import { useToast } from '@/components/ui/use-toast';
import {
  Job,
  PauseJobRequest,
  PauseJobResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';

const FORM_SCHEMA = Yup.object({
  note: Yup.string().optional(),
  isPaused: Yup.boolean().required(),
  error: Yup.string().optional(),
});

export type FormValues = Yup.InferType<typeof FORM_SCHEMA>;

interface Props {
  job: Job;
  mutate: () => void;
}

export default function JobPauseSwitch({ job, mutate }: Props): ReactElement {
  const { toast } = useToast();
  const form = useForm({
    resolver: yupResolver<FormValues>(FORM_SCHEMA),
    defaultValues: {
      isPaused: job?.pauseStatus?.isPaused || false,
      note: job?.pauseStatus?.note || '',
      error: job.pauseStatus ? '' : 'Unable to retrieve job pause status',
    },
  });

  async function onSubmit(values: FormValues) {
    try {
      await pauseJob(job.id, values.isPaused, values.note);
      toast({
        title: `Successfully ${values.isPaused ? 'paused' : 'unpaused'}  job!`,
        variant: 'default',
      });
      mutate();
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to pause',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Pause</CardTitle>
      </CardHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)}>
          <CardContent>
            {form.getValues().error != '' ? (
              <Alert variant="destructive">
                <AlertTitle>{`Error: ${form.getValues().error}`}</AlertTitle>
              </Alert>
            ) : (
              <div className="flex flex-row  space-x-4">
                <div className="basis-1/3">
                  <FormField
                    control={form.control}
                    name="isPaused"
                    render={({ field }) => (
                      <FormItem>
                        <FormControl>
                          <SwitchCard
                            isChecked={field.value || false}
                            onCheckedChange={field.onChange}
                            title="Pause job"
                            description="Prevents future job runs."
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                <div className="basis-2/3">
                  <FormField
                    control={form.control}
                    name="note"
                    render={({ field }) => (
                      <FormItem>
                        <FormControl>
                          <Textarea
                            className="h-[75px]"
                            placeholder="Type a note here."
                            {...field}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
              </div>
            )}
          </CardContent>
          <CardFooter className="bg-muted">
            <div className="flex flex-row items-center justify-end w-full mt-4">
              <Button type="submit" disabled={!form.formState.isDirty}>
                Save
              </Button>
            </div>
          </CardFooter>
        </form>
      </Form>
    </Card>
  );
}

async function pauseJob(
  jobId: string,
  isPaused: boolean,
  note?: string
): Promise<PauseJobResponse> {
  const res = await fetch(`/api/jobs/${jobId}/pause`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new PauseJobRequest({
        id: jobId,
        pause: isPaused,
        note: note,
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return PauseJobResponse.fromJson(await res.json());
}
