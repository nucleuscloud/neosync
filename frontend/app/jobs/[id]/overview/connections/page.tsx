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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';
import { getJob } from '../util';

const CONNECTIONS_FORM_SCHEMA = Yup.object({
  sourceId: Yup.string().uuid().required(),
  destinationId: Yup.string().uuid().required(),
});

export type ConnectionsFormValues = Yup.InferType<
  typeof CONNECTIONS_FORM_SCHEMA
>;

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';

  const { isLoading: isConnectionsLoading, data: connectionsData } =
    useGetConnections();

  const connections = connectionsData?.connections ?? [];

  const form = useForm({
    resolver: yupResolver<ConnectionsFormValues>(CONNECTIONS_FORM_SCHEMA),
    defaultValues: async () => {
      const res = await getJob(id);
      if (!res) {
        return { sourceId: '', destinationId: '' };
      }
      return {
        sourceId: res.job?.connectionSourceId || '',
        destinationId: res.job?.connectionDestinationIds[0] || '',
      };
    },
  });

  async function onSubmit(_values: ConnectionsFormValues) {}

  return (
    <div className="job-details-container">
      <PageHeader header="Connections" description="Manage job connections" />
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <FormField
            control={form.control}
            name="sourceId"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Source</FormLabel>
                <FormControl>
                  {isConnectionsLoading ? (
                    <Skeleton />
                  ) : (
                    <Select onValueChange={field.onChange} value={field.value}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {connections
                          .filter(
                            (c) => c.id !== form.getValues().destinationId
                          )
                          .map((connection) => (
                            <SelectItem
                              className="cursor-pointer"
                              key={connection.id}
                              value={connection.id}
                            >
                              {connection.name}
                            </SelectItem>
                          ))}
                      </SelectContent>
                    </Select>
                  )}
                </FormControl>
                <FormDescription>
                  The location of the source data set.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="destinationId"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Destination</FormLabel>
                <FormControl>
                  {isConnectionsLoading ? (
                    <Skeleton />
                  ) : (
                    <Select onValueChange={field.onChange} value={field.value}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {connections
                          .filter((c) => c.id !== form.getValues().sourceId)
                          .map((connection) => (
                            <SelectItem
                              className="cursor-pointer"
                              key={connection.id}
                              value={connection.id}
                            >
                              {connection.name}
                            </SelectItem>
                          ))}
                      </SelectContent>
                    </Select>
                  )}
                </FormControl>
                <FormDescription>
                  The location of the destination data set.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <div className="flex flex-row gap-1 justify-end">
            <Button type="submit">Save</Button>
          </div>
        </form>
      </Form>
    </div>
  );
}
