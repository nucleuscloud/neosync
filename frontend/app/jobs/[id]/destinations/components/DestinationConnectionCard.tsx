'use client';
import DestinationOptionsForm from '@/components/jobs/Form/DestinationOptionsForm';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
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
  JobDestination,
  JobDestinationOptions,
  SqlDestinationConnectionOptions,
  UpdateJobDestinationConnectionRequest,
  UpdateJobDestinationConnectionResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { getErrorMessage } from '@/util/util';
import { getConnectionType } from '@/yup-validations/connections';
import { DESTINATION_FORM_SCHEMA } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { Cross2Icon, PlusIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
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
  const account = useAccount();
  const {
    isLoading: isConnectionsLoading,
    data: connectionsData,
    mutate,
  } = useGetConnections(account?.id ?? '');

  const connections = connectionsData?.connections ?? [];

  const form = useForm({
    resolver: yupResolver<FormValues>(FORM_SCHEMA),
    defaultValues: async () => {
      const res = await getJob(jobId);
      if (!res || !res.job?.destinations) {
        return {
          sourceId: '',
          destinations: [
            { destinationId: '', destinationOptions: {}, type: '' },
          ],
        };
      }
      const destinations = res.job?.destinations.map((d) => {
        switch (d.options?.config.case) {
          case 'sqlOptions':
            return {
              destinationId: d.connectionId,
              destinationOptions: {
                truncateBeforeInsert:
                  d.options.config.value.truncateBeforeInsert,
                initDbSchema: d.options.config.value.initDbSchema,
              },
              type: 'sql',
            };
          default:
            return {
              destinationId: d.connectionId,
              destinationOptions: {},
              type: '',
            };
        }
      });

      return {
        sourceId: res.job?.source?.connectionId || '',
        destinations: destinations,
      };
    },
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: 'destinations',
  });

  async function onSubmit(values: FormValues) {
    try {
      await updateJobConnections(jobId, values);
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

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        <div className="space-y-12">
          {fields.map(({}, index) => {
            return (
              <div className="space-y-4" key={index}>
                <div className="flex flew-row justify-between">
                  <div className="basis-11/12">
                    <FormField
                      control={form.control}
                      name={`destinations.${index}.destinationId`}
                      render={({ field }) => (
                        <FormItem>
                          <FormControl>
                            {isConnectionsLoading ? (
                              <Skeleton />
                            ) : (
                              <Select
                                onValueChange={(value: string) => {
                                  field.onChange(value);
                                  form.setValue(
                                    `destinations.${index}.type`,
                                    getConnectionType(
                                      connections.find((c) => c.id == value)
                                    )
                                  );
                                  form.setValue(
                                    `destinations.${index}.destinationOptions`,
                                    {
                                      truncateBeforeInsert: false,
                                      initDbSchema: false,
                                    }
                                  );
                                }}
                                value={field.value}
                              >
                                <SelectTrigger>
                                  <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                  {connections
                                    .filter(
                                      (c) => c.id !== form.getValues().sourceId
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
                            The location of the destination data set.
                          </FormDescription>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      if (fields.length != 1) {
                        remove(index);
                      }
                    }}
                  >
                    <Cross2Icon className="w-4 h-4" />
                  </Button>
                </div>
                <DestinationOptionsForm
                  index={index}
                  connection={connections.find(
                    (c) =>
                      c.id == form.getValues().destinations[index].destinationId
                  )}
                  maxColNum={2}
                />
              </div>
            );
          })}
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              append({
                destinationId: '',
                destinationOptions: {},
                type: '',
              });
            }}
          >
            Add
            <PlusIcon className="ml-2 w-4 h-4" />
          </Button>
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
  values: FormValues
): Promise<UpdateJobDestinationConnectionResponse> {
  const res = await fetch(`/api/jobs/${jobId}/destination-connection`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new UpdateJobDestinationConnectionRequest({
        id: jobId,
        destination: new JobDestination({
          connectionId: values.destinationId,
          options: new JobDestinationOptions({
            config: {
              case: 'sqlOptions',
              value: new SqlDestinationConnectionOptions({
                truncateBeforeInsert:
                  values.destinationOptions.truncateBeforeInsert,
                initDbSchema: values.destinationOptions.initDbSchema,
              }),
            },
          }),
        }),
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateJobDestinationConnectionResponse.fromJson(await res.json());
}
