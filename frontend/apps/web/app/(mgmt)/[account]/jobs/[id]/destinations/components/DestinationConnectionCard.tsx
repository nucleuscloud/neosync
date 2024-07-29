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
import { getErrorMessage } from '@/util/util';
import { NewDestinationFormValues } from '@/yup-validations/jobs';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  Connection,
  JobDestination,
  JobDestinationOptions,
} from '@neosync/sdk';
import {
  deleteJobDestinationConnection,
  updateJobDestinationConnection,
} from '@neosync/sdk/connectquery';
import { ReactElement } from 'react';
import { Control, useForm, useWatch } from 'react-hook-form';
import {
  getDefaultDestinationFormValues,
  toJobDestinationOptions,
} from '../../../util';

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
  const { mutateAsync: setJobDestConnection } = useMutation(
    updateJobDestinationConnection
  );
  const { mutateAsync: removeJobDestConnection } = useMutation(
    deleteJobDestinationConnection
  );

  const form = useForm({
    resolver: yupResolver<DestinationFormValues>(NewDestinationFormValues),
    values: getDefaultDestinationFormValues(destination),
  });

  async function onSubmit(values: DestinationFormValues) {
    try {
      const connection = connections.find((c) => c.id === values.connectionId);
      await setJobDestConnection({
        jobId,
        connectionId: values.connectionId,
        destinationId: destination.id,
        options: new JobDestinationOptions(
          toJobDestinationOptions(values, connection)
        ),
      });
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
      await removeJobDestConnection({
        destinationId: destination.id,
      });
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

function useShouldHideInitConnectionSchema(
  control: Control<DestinationFormValues>,
  sourceId: string
): boolean {
  const [destinationConnectionid] = useWatch({
    control,
    name: ['connectionId'],
  });
  return destinationConnectionid === sourceId;
}
