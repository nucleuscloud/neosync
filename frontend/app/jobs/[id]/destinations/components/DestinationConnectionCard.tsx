'use client';
import DestinationOptionsForm from '@/components/jobs/Form/DestinationOptionsForm';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter } from '@/components/ui/card';
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
import { useToast } from '@/components/ui/use-toast';
import { Connection } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import {
  JobDestination,
  JobDestinationOptions,
  SetJobDestinationConnectionRequest,
  SetJobDestinationConnectionResponse,
  SqlDestinationConnectionOptions,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import { DESTINATION_FORM_SCHEMA } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';

export const FORM_SCHEMA = DESTINATION_FORM_SCHEMA;
export type FormValues = Yup.InferType<typeof FORM_SCHEMA>;

interface Props {
  jobId: string;
  destination: JobDestination;
  connections: Connection[];
  availableConnections: Connection[];
  mutate: () => {};
}

export default function DestinationConnectionCard({
  jobId,
  destination,
  connections,
  availableConnections,
  mutate,
}: Props): ReactElement {
  const { toast } = useToast();

  const connection = connections.find((c) => c.id == destination.connectionId);
  const form = useForm({
    resolver: yupResolver<FormValues>(FORM_SCHEMA),
    defaultValues: getDefaultValues(destination),
  });

  async function onSubmit(values: FormValues) {
    try {
      await setJobConnection(jobId, values, connection);
      mutate();
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

  async function onDelete() {
    try {
      await deleteJobConnection(jobId, destination.connectionId);
      mutate();
      toast({
        title: 'Successfully deleted job destination!',
        variant: 'default',
      });
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to delete job destination',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <Card>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)}>
          <CardContent className="mt-6">
            <div className="space-y-4">
              <FormField
                control={form.control}
                name="destinationId"
                render={({ field }) => (
                  <FormItem>
                    <FormControl>
                      <Select
                        onValueChange={(value: string) => {
                          field.onChange(value);
                          form.setValue(`destinationOptions`, {
                            truncateBeforeInsert: false,
                            initDbSchema: false,
                          });
                        }}
                        value={field.value}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {availableConnections.map((connection) => (
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
                    </FormControl>
                    <FormDescription>
                      The location of the destination data set.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <DestinationOptionsForm
                connection={connections.find(
                  (c) => c.id == form.getValues().destinationId
                )}
                maxColNum={2}
              />
            </div>
          </CardContent>
          <CardFooter>
            <div className="flex flex-row items-center justify-between w-full mt-4">
              <Button type="button" variant="destructive" onClick={onDelete}>
                Delete
              </Button>
              <Button disabled={!form.formState.isDirty} type="submit">
                Save
              </Button>
            </div>
          </CardFooter>
        </form>
      </Form>
    </Card>
  );
}

async function deleteJobConnection(
  jobId: string,
  connectionId: string
): Promise<void> {
  const res = await fetch(
    `/api/jobs/${jobId}/destination-connection/${connectionId}`,
    {
      method: 'DELETE',
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}

async function setJobConnection(
  jobId: string,
  values: FormValues,
  connection?: Connection
): Promise<SetJobDestinationConnectionResponse> {
  const res = await fetch(`/api/jobs/${jobId}/destination-connection`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new SetJobDestinationConnectionRequest({
        jobId: jobId,
        connectionId: values.destinationId,
        options: new JobDestinationOptions(
          toJobDestinationOptions(values, connection)
        ),
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return SetJobDestinationConnectionResponse.fromJson(await res.json());
}

function toJobDestinationOptions(
  values: FormValues,
  connection?: Connection
): JobDestinationOptions {
  if (!connection) {
    return new JobDestinationOptions();
  }
  switch (connection.connectionConfig?.config.case) {
    case 'pgConfig': {
      return new JobDestinationOptions({
        config: {
          case: 'sqlOptions',
          value: new SqlDestinationConnectionOptions({
            truncateBeforeInsert:
              values.destinationOptions.truncateBeforeInsert,
            initDbSchema: values.destinationOptions.initDbSchema,
          }),
        },
      });
    }
    default: {
      return new JobDestinationOptions();
    }
  }
}

function getDefaultValues(d: JobDestination): FormValues {
  switch (d.options?.config.case) {
    case 'sqlOptions':
      return {
        destinationId: d.connectionId,
        destinationOptions: {
          truncateBeforeInsert: d.options.config.value.truncateBeforeInsert,
          initDbSchema: d.options.config.value.initDbSchema,
        },
      };
    default:
      return {
        destinationId: d.connectionId,
        destinationOptions: {},
      };
  }
}
