'use client';
import DestinationOptionsForm from '@/components/jobs/Form/DestinationOptionsForm';
import { useAccount } from '@/components/providers/account-provider';
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
import { getErrorMessage } from '@/util/util';
import {
  DESTINATION_FORM_SCHEMA,
  toJobDestinationOptions,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  Connection,
  JobDestination,
  JobDestinationOptions,
  UpdateJobDestinationConnectionRequest,
  UpdateJobDestinationConnectionResponse,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { Control, useForm, useWatch } from 'react-hook-form';
import * as Yup from 'yup';

const FORM_SCHEMA = DESTINATION_FORM_SCHEMA;
type FormValues = Yup.InferType<typeof FORM_SCHEMA>;

interface Props {
  jobId: string;
  jobSourceId: string;
  destination: JobDestination;
  connections: Connection[];
  availableConnections: Connection[];
  mutate: () => {};
  isDeleteDisabled?: boolean;
}

export default function DestinationConnectionCard({
  jobId,
  destination,
  connections,
  availableConnections,
  mutate,
  isDeleteDisabled,
  jobSourceId,
}: Props): ReactElement {
  const { toast } = useToast();
  const { account } = useAccount();

  const form = useForm({
    resolver: yupResolver<FormValues>(FORM_SCHEMA),
    values: getDefaultValues(destination),
  });

  async function onSubmit(values: FormValues) {
    try {
      const connection = connections.find((c) => c.id === values.connectionId);
      await setJobConnection(
        account?.id ?? '',
        jobId,
        values,
        destination.id,
        connection
      );
      mutate();
      toast({
        title: 'Successfully updated job destination!',
        variant: 'success',
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
      await deleteJobConnection(account?.id ?? '', jobId, destination.id);
      mutate();
      toast({
        title: 'Successfully deleted job destination!',
        variant: 'success',
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

  const dest = availableConnections.find(
    (item) => item.id === destination.connectionId
  );
  const destOpts = form.watch('destinationOptions');
  const shouldHideInitTableSchema = useShouldHideInitConnectionSchema(
    form.control,
    jobSourceId
  );
  return (
    <Card>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)}>
          <CardContent className="mt-6">
            <div className="space-y-4">
              <FormField
                control={form.control}
                name="connectionId"
                render={({ field }) => (
                  <FormItem>
                    <FormControl>
                      <Select
                        onValueChange={(value: string) => {
                          field.onChange(value);
                          form.setValue(
                            `destinationOptions`,
                            {
                              truncateBeforeInsert: false,
                              truncateCascade: false,
                              initTableSchema: false,
                              onConflictDoNothing: false,
                            },
                            {
                              shouldDirty: true,
                              shouldTouch: true,
                              shouldValidate: true,
                            }
                          );
                        }}
                        value={field.value}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder={dest?.name} />
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
                  (c) => c.id === form.getValues().connectionId
                )}
                value={{
                  initTableSchema: destOpts.initTableSchema ?? false,
                  onConflictDoNothing: destOpts.onConflictDoNothing ?? false,
                  truncateBeforeInsert: destOpts.truncateBeforeInsert ?? false,
                  truncateCascade: destOpts.truncateCascade ?? false,
                }}
                setValue={(newOpts) => {
                  form.setValue(
                    'destinationOptions',
                    {
                      initTableSchema: newOpts.initTableSchema,
                      onConflictDoNothing: newOpts.onConflictDoNothing,
                      truncateBeforeInsert: newOpts.truncateBeforeInsert,
                      truncateCascade: newOpts.truncateCascade,
                    },
                    {
                      shouldDirty: true,
                      shouldTouch: true,
                      shouldValidate: true,
                    }
                  );
                }}
                hideInitTableSchema={shouldHideInitTableSchema}
              />
            </div>
          </CardContent>
          <CardFooter>
            <div className="flex flex-row items-center justify-between w-full mt-4">
              <Button
                type="button"
                variant="destructive"
                onClick={onDelete}
                disabled={isDeleteDisabled}
              >
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
  accountId: string,
  jobId: string,
  destinationId: string
): Promise<void> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${jobId}/destination-connection/${destinationId}`,
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
  accountId: string,
  jobId: string,
  values: FormValues,
  destinationId: string,
  connection?: Connection
): Promise<UpdateJobDestinationConnectionResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${jobId}/destination-connection`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(
        new UpdateJobDestinationConnectionRequest({
          jobId: jobId,
          connectionId: values.connectionId,
          destinationId: destinationId,
          options: new JobDestinationOptions(
            toJobDestinationOptions(values, connection)
          ),
        })
      ),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateJobDestinationConnectionResponse.fromJson(await res.json());
}

function getDefaultValues(d: JobDestination): FormValues {
  switch (d.options?.config.case) {
    case 'postgresOptions':
      return {
        connectionId: d.connectionId,
        destinationOptions: {
          truncateBeforeInsert:
            d.options.config.value.truncateTable?.truncateBeforeInsert,
          truncateCascade: d.options.config.value.truncateTable?.cascade,
          initTableSchema: d.options.config.value.initTableSchema,
          onConflictDoNothing: d.options.config.value.onConflict?.doNothing,
        },
      };
    case 'mysqlOptions':
      return {
        connectionId: d.connectionId,
        destinationOptions: {
          truncateBeforeInsert:
            d.options.config.value.truncateTable?.truncateBeforeInsert,
          initTableSchema: d.options.config.value.initTableSchema,
          onConflictDoNothing: d.options.config.value.onConflict?.doNothing,
        },
      };
    default:
      return {
        connectionId: d.connectionId,
        destinationOptions: {},
      };
  }
}

function useShouldHideInitConnectionSchema(
  control: Control<FormValues>,
  sourceId: string
): boolean {
  const [destinationConnectionid] = useWatch({
    control,
    name: ['connectionId'],
  });
  return destinationConnectionid === sourceId;
}
