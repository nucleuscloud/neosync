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
import { yupResolver } from '@hookform/resolvers/yup';
import cron from 'cron-validate';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';
import { getJob } from '../util';

const SCHEDULE_FORM_SCHEMA = Yup.object({
  cronSchedule: Yup.string()
    .optional()
    .test('isValidCron', 'Not a valid cron schedule', (value) => {
      return !!value && cron(value).isValid();
    }),
});

export type ScheduleFormValues = Yup.InferType<typeof SCHEDULE_FORM_SCHEMA>;

interface Props {
  jobId: string;
}

export default function JobScheduleCard({ jobId }: Props): ReactElement {
  const form = useForm({
    resolver: yupResolver<ScheduleFormValues>(SCHEDULE_FORM_SCHEMA),
    defaultValues: async () => {
      const res = await getJob(jobId);
      if (!res) {
        return { cronSchedule: '' };
      }
      return { cronSchedule: res.job?.cronSchedule };
    },
  });

  async function onSubmit(_values: ScheduleFormValues) {}

  return (
    <Card>
      <CardHeader>
        <CardTitle>Schedule</CardTitle>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
            <FormField
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
            />
          </form>
        </Form>
      </CardContent>
      <CardFooter className="bg-muted">
        <div className="flex flex-row items-center justify-between w-full mt-4">
          <p className="text-muted-foreground"></p>
          <Button type="submit">Save</Button>
        </div>
      </CardFooter>
    </Card>
  );
}
