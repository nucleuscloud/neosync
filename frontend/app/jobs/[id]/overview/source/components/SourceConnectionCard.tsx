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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { useToast } from '@/components/ui/use-toast';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import {
  UpdateJobSourceConnectionRequest,
  UpdateJobSourceConnectionResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';
import { getJob } from '../../util';

const CONNECTIONS_FORM_SCHEMA = Yup.object({
  sourceId: Yup.string().uuid().required(),
  destinationId: Yup.string().uuid().required(),
});

export type ConnectionsFormValues = Yup.InferType<
  typeof CONNECTIONS_FORM_SCHEMA
>;

interface Props {
  jobId: string;
}

export default function SourceConnectionCard({ jobId }: Props): ReactElement {
  const { toast } = useToast();
  const { isLoading: isConnectionsLoading, data: connectionsData } =
    useGetConnections();

  const connections = connectionsData?.connections ?? [];

  const form = useForm({
    resolver: yupResolver<ConnectionsFormValues>(CONNECTIONS_FORM_SCHEMA),
    defaultValues: async () => {
      const res = await getJob(jobId);
      if (!res) {
        return { sourceId: '', destinationId: '' };
      }
      return {
        sourceId: res.job?.connectionSourceId || '',
        destinationId: res.job?.connectionDestinationIds[0] || '',
      };
    },
  });

  async function onSubmit(values: ConnectionsFormValues) {
    try {
      await updateJobConnection(jobId, values.sourceId);
      toast({
        title: 'Successfully updated job schedule!',
        variant: 'default',
      });
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to update job schedule',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Source</CardTitle>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <FormField
              control={form.control}
              name="sourceId"
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    {isConnectionsLoading ? (
                      <Skeleton />
                    ) : (
                      <Select
                        onValueChange={field.onChange}
                        value={field.value}
                      >
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
          </form>
        </Form>
      </CardContent>
      <CardFooter className="bg-muted">
        <div className="flex flex-row items-center justify-between w-full mt-4">
          <p className="text-muted-foreground text-sm">
            It may take a minute to validate your connection
          </p>
          <Button type="submit">Save</Button>
        </div>
      </CardFooter>
    </Card>
  );
}

async function updateJobConnection(
  jobId: string,
  connectionId: string
): Promise<UpdateJobSourceConnectionResponse> {
  const res = await fetch(`/api/jobs/${jobId}/schedule`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new UpdateJobSourceConnectionRequest({
        id: jobId,
        connectionId: connectionId,
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateJobSourceConnectionResponse.fromJson(await res.json());
}
