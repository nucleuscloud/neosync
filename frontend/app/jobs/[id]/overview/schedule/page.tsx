'use client';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
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

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';

  const form = useForm({
    resolver: yupResolver<ScheduleFormValues>(SCHEDULE_FORM_SCHEMA),
    defaultValues: async () => {
      const res = await getJob(id);
      if (!res) {
        return { cronSchedule: '' };
      }
      return { cronSchedule: res.job?.cronSchedule };
    },
  });

  async function onSubmit(_values: ScheduleFormValues) {}

  return (
    <div className="job-details-container">
      <PageHeader header="Schedule" description="Manage job schedule" />
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <FormField
            control={form.control}
            name="cronSchedule"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Cron Schedule</FormLabel>
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

          <div className="flex flex-row justify-end">
            <Button type="submit">Save</Button>
          </div>
        </form>
      </Form>
    </div>
  );
}
