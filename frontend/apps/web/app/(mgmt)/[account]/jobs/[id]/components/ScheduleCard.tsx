'use client';
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
  FormDescription,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { Job, JobService } from '@neosync/sdk';
import cron from 'cron-validate';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import * as Yup from 'yup';

export const DEFAULT_CRON_STRING = '0 0 1 1 *';

const SCHEDULE_FORM_SCHEMA = Yup.object({
  cronSchedule: Yup.string()
    .optional()
    .test('isValidCron', 'Not a valid cron schedule', (value, context) => {
      if (!value) {
        return false;
      }
      const output = cron(value);
      if (output.isValid()) {
        return true;
      }
      if (output.isError()) {
        const errors = output.getError();
        if (errors.length > 0) {
          return context.createError({ message: errors.join(', ') });
        }
      }
      return output.isValid();
    }),
});

type ScheduleFormValues = Yup.InferType<typeof SCHEDULE_FORM_SCHEMA>;

interface Props {
  job: Job;
  mutate: (newjob: Job) => void;
}

export default function JobScheduleCard({ job, mutate }: Props): ReactElement {
  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver<ScheduleFormValues>(SCHEDULE_FORM_SCHEMA),
    values: { cronSchedule: job?.cronSchedule },
  });
  const { mutateAsync: updateJobScheduleAsync } = useMutation(
    JobService.method.updateJobSchedule
  );

  async function onSubmit(values: ScheduleFormValues) {
    try {
      const resp = await updateJobScheduleAsync({
        id: job.id,
        cronSchedule: values.cronSchedule,
      });
      toast.success('Successfully updated job schedule!');
      if (resp.job) {
        mutate(resp.job);
      }
    } catch (err) {
      console.error(err);
      toast.error('Unable to update job schedule', {
        description: getErrorMessage(err),
      });
    }
  }

  const msg =
    !form.getValues().cronSchedule || form.getValues().cronSchedule === ''
      ? 'Not currently scheduled'
      : '';

  return (
    <Card className="overflow-hidden">
      <CardHeader>
        <CardTitle>Schedule</CardTitle>
      </CardHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)}>
          <CardContent className="p-5">
            <FormField
              control={form.control}
              name="cronSchedule"
              render={({ field }) => (
                <FormItem>
                  <FormDescription>
                    The schedule to run the job against if not a oneoff.
                  </FormDescription>
                  <FormControl>
                    <Input
                      placeholder="Cron Schedule"
                      value={field.value || ''}
                      onChange={field.onChange}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </CardContent>
          <CardFooter className="bg-muted flex py-2 justify-center">
            <div className="flex flex-row items-center justify-between w-full">
              <p className="text-muted-foreground text-sm">{msg}</p>
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
