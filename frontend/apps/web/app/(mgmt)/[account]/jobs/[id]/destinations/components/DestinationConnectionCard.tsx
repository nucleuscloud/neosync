'use client';
import ConnectionSelectContent from '@/app/(mgmt)/[account]/new/job/connect/ConnectionSelectContent';
import ButtonText from '@/components/ButtonText';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import DestinationOptionsForm from '@/components/jobs/Form/DestinationOptionsForm';
import { useAccount } from '@/components/providers/account-provider';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Alert } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter } from '@/components/ui/card';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
} from '@/components/ui/form';
import {
  Select,
  SelectContent,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { splitConnections } from '@/libs/utils';
import { getErrorMessage } from '@/util/util';
import { NewDestinationFormValues } from '@/yup-validations/jobs';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  Connection,
  JobDestination,
  JobMapping,
  JobService,
} from '@neosync/sdk';
import { TrashIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import { Control, useForm, useWatch } from 'react-hook-form';
import { IoAlertCircleOutline } from 'react-icons/io5';
import { toast } from 'sonner';
import {
  getDestinationFormValuesOrDefaultFromDestination,
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
  jobmappings?: JobMapping[];
}

export default function DestinationConnectionCard({
  jobId,
  destination,
  connections,
  availableConnections,
  mutate,
  isDeleteDisabled,
  jobSourceId,
  jobmappings,
}: Props): ReactElement {
  const { account } = useAccount();
  const { mutateAsync: setJobDestConnection } = useMutation(
    JobService.method.updateJobDestinationConnection
  );
  const { mutateAsync: removeJobDestConnection } = useMutation(
    JobService.method.deleteJobDestinationConnection
  );

  const form = useForm({
    resolver: yupResolver<NewDestinationFormValues>(NewDestinationFormValues),
    values: getDestinationFormValuesOrDefaultFromDestination(destination),
  });

  const { data: validateSchemaResponse, isLoading: isValidatingSchema } =
    useQuery(
      JobService.method.validateSchema,
      {
        connectionId: form.getValues('connectionId'),
        mappings: jobmappings,
      },
      {
        enabled: !!jobmappings && !!form.getValues('connectionId'),
      }
    );

  async function onSubmit(values: NewDestinationFormValues) {
    try {
      const connection = connections.find((c) => c.id === values.connectionId);
      await setJobDestConnection({
        jobId,
        connectionId: values.connectionId,
        destinationId: destination.id,
        options: toJobDestinationOptions(values, connection),
      });
      mutate();
      toast.success('Successfully updated job destination!');
    } catch (err) {
      console.error(err);
      toast.error('Unable to update job destination', {
        description: getErrorMessage(err),
      });
    }
  }

  async function onDelete() {
    try {
      await removeJobDestConnection({
        destinationId: destination.id,
      });
      mutate();
      toast.success('Successfully deleted job destination!');
    } catch (err) {
      console.error(err);
      toast.error('Unable to delete job destination', {
        description: getErrorMessage(err),
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

  const columnErrors =
    !!validateSchemaResponse?.missingColumns?.length ||
    !!validateSchemaResponse?.extraColumns?.length;

  const tableErrors =
    !!validateSchemaResponse?.missingTables?.length ||
    !!validateSchemaResponse?.missingSchemas?.length;

  const { postgres, mysql, s3, mongodb, gcpcs, dynamodb, mssql } =
    splitConnections(availableConnections);
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
                            {},
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
                          <SelectValue
                            ref={field.ref}
                            placeholder={dest?.name}
                          />
                        </SelectTrigger>
                        <SelectContent>
                          <ConnectionSelectContent
                            postgres={postgres}
                            mysql={mysql}
                            s3={s3}
                            mongodb={mongodb}
                            gcpcs={gcpcs}
                            dynamodb={dynamodb}
                            mssql={mssql}
                          />
                        </SelectContent>
                      </Select>
                    </FormControl>
                    <FormDescription>
                      The location of the destination data set.
                    </FormDescription>
                  </FormItem>
                )}
              />
              {isValidatingSchema && (
                <Skeleton className="w-full h-24 rounded-lg" />
              )}
              {!isValidatingSchema && columnErrors && (
                <div>
                  <Alert className="border-red-400">
                    <Accordion type="single" collapsible className="w-full">
                      <AccordionItem value={`table-error`} key={1}>
                        <AccordionTrigger className="text-left">
                          <div className="font-medium flex flex-row items-center gap-2">
                            <IoAlertCircleOutline className="h-6 w-6" />
                            Found issues with columns in schema. Please resolve
                            before next job run.
                          </div>
                        </AccordionTrigger>
                        <AccordionContent>
                          <div className="space-y-2">
                            <div className="rounded-md bg-muted p-3">
                              <pre className="mt-2 whitespace-pre-wrap text-sm">
                                {(() => {
                                  const { missingColumns, extraColumns } =
                                    validateSchemaResponse || {};
                                  let output = '';
                                  if (
                                    missingColumns &&
                                    missingColumns.length > 0
                                  ) {
                                    output +=
                                      'Columns Missing in Destination:\n';
                                    missingColumns.forEach((col) => {
                                      output += ` - ${col.schema}.${col.table}.${col.column}\n`;
                                    });
                                    output += '\n';
                                  }
                                  if (extraColumns && extraColumns.length > 0) {
                                    output +=
                                      'Columns Not Found in Job Mappings:\n';
                                    extraColumns.forEach((col) => {
                                      output += ` - ${col.schema}.${col.table}.${col.column}\n`;
                                    });
                                  }
                                  return output || 'No differences found.';
                                })()}
                              </pre>
                            </div>
                          </div>
                        </AccordionContent>
                      </AccordionItem>
                    </Accordion>
                  </Alert>
                </div>
              )}
              {!isInitTableSchemaEnabled(
                form.getValues('destinationOptions')
              ) &&
                !isValidatingSchema &&
                tableErrors && (
                  <div>
                    <Alert className="border-red-400">
                      <Accordion type="single" collapsible className="w-full">
                        <AccordionItem value={`table-error`} key={1}>
                          <AccordionTrigger className="text-left">
                            <div className="font-medium flex flex-row items-center gap-2">
                              <IoAlertCircleOutline className="h-6 w-6" />
                              This destination is missing tables found in Job
                              Mappings. Either enable Init Table Schema or
                              create the tables manually.
                            </div>
                          </AccordionTrigger>
                          <AccordionContent>
                            <div className="space-y-2">
                              <div className="rounded-md bg-muted p-3">
                                <pre className="mt-2 whitespace-pre-wrap text-sm">
                                  {(() => {
                                    const { missingSchemas, missingTables } =
                                      validateSchemaResponse || {};
                                    let output = '';
                                    if (
                                      missingSchemas &&
                                      missingSchemas.length > 0
                                    ) {
                                      output +=
                                        'Schemas Missing in Destination:\n';
                                      missingSchemas.forEach((schema) => {
                                        output += ` - ${schema}\n`;
                                      });
                                      output += '\n';
                                    }
                                    if (
                                      missingTables &&
                                      missingTables.length > 0
                                    ) {
                                      output +=
                                        'Tables Missing in Destination:\n';
                                      missingTables.forEach((table) => {
                                        output += ` - ${table.schema}.${table.table}\n`;
                                      });
                                      output += '\n';
                                    }
                                    return output || 'No differences found.';
                                  })()}
                                </pre>
                              </div>
                            </div>
                          </AccordionContent>
                        </AccordionItem>
                      </Accordion>
                    </Alert>
                  </div>
                )}

              <DestinationOptionsForm
                connection={connections.find(
                  (c) => c.id === form.getValues().connectionId
                )}
                value={destOpts}
                setValue={(newOpts) => {
                  form.setValue('destinationOptions', newOpts, {
                    shouldDirty: true,
                    shouldTouch: true,
                    shouldValidate: true,
                  });
                }}
                hideInitTableSchema={shouldHideInitTableSchema}
                hideDynamoDbTableMappings={true}
                destinationDetailsRecord={{}} // not used because we are hiding dynamodb table mappings
                errors={form.formState.errors?.destinationOptions}
              />
            </div>
          </CardContent>
          <CardFooter>
            <div className="flex flex-row items-center justify-between w-full mt-4">
              <DeleteButton isDisabled={isDeleteDisabled} onDelete={onDelete} />
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

function isInitTableSchemaEnabled(
  destinationOptions: NewDestinationFormValues['destinationOptions']
): boolean {
  return (
    destinationOptions?.postgres?.initTableSchema ||
    destinationOptions?.mssql?.initTableSchema ||
    destinationOptions?.mysql?.initTableSchema ||
    false
  );
}

interface DeleteButtonProps {
  isDisabled?: boolean;
  onDelete(): Promise<void> | void;
}

function DeleteButton(props: DeleteButtonProps): ReactElement {
  const { isDisabled, onDelete } = props;
  return (
    <DeleteConfirmationDialog
      trigger={
        <Button type="button" variant="destructive" disabled={isDisabled}>
          <ButtonText leftIcon={<TrashIcon />} text="Delete" />
        </Button>
      }
      headerText="Are you sure you want to delete this destination connection?"
      description="Deleting this is irreversable and will cause data to stop syncing to this destination!"
      onConfirm={async () => onDelete()}
    />
  );
}

function useShouldHideInitConnectionSchema(
  control: Control<NewDestinationFormValues>,
  sourceId: string
): boolean {
  const [destinationConnectionid] = useWatch({
    control,
    name: ['connectionId'],
  });
  return destinationConnectionid === sourceId;
}
