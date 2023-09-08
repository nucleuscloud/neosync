'use client';
import DestinationOptionsForm from '@/components/jobs/Form/DestinationOptionsForm';
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
import { useToast } from '@/components/ui/use-toast';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import {
  UpdateJobDestinationConnectionsRequest,
  UpdateJobDestinationConnectionsResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import { DESTINATION_FORM_SCHEMA } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';
import { getJob } from '../../util';

export const FORM_SCHEMA = DESTINATION_FORM_SCHEMA.concat(
  Yup.object({
    sourceId: Yup.string().required(),
  })
);
export type FormValues = Yup.InferType<typeof FORM_SCHEMA>;

interface Props {
  jobId: string;
}

export default function DestinationConnectionCard({
  jobId,
}: Props): ReactElement {
  const { toast } = useToast();
  const { isLoading: isConnectionsLoading, data: connectionsData } =
    useGetConnections();

  const connections = connectionsData?.connections ?? [];

  const form = useForm({
    resolver: yupResolver<FormValues>(FORM_SCHEMA),
    defaultValues: async () => {
      const res = await getJob(jobId);
      if (!res) {
        return { sourceId: '', destinationOptions: {}, destinationId: '' };
      }
      const destinationIds = res.job?.destinations.map((d) => d.connectionId);
      return {
        sourceId: res.job?.source?.connectionId || '',
        destinationOptions: {},
        destinationId: destinationIds ? destinationIds[0] : '',
      };
    },
  });

  async function onSubmit(values: FormValues) {
    try {
      await updateJobConnections(jobId, [values.destinationId]);
      toast({
        title: 'Successfully updated job destination!',
        variant: 'default',
      });
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to update job destination',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        <div className="space-y-8">
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

          <DestinationOptionsForm
            formControl={form.control}
            connection={connections.find(
              (c) => c.id == form.getValues().destinationId
            )}
            maxColNum={2}
          />
          <div className="flex flex-row items-center justify-end w-full mt-4">
            <Button disabled={!form.formState.isDirty} type="submit">
              Save
            </Button>
          </div>
        </div>
      </form>
    </Form>
  );
}

async function updateJobConnections(
  jobId: string,
  connectionIds: string[]
): Promise<UpdateJobDestinationConnectionsResponse> {
  const res = await fetch(`/api/jobs/${jobId}/destination-connections`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new UpdateJobDestinationConnectionsRequest({
        id: jobId,
        connectionIds: connectionIds,
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateJobDestinationConnectionsResponse.fromJson(await res.json());
}
