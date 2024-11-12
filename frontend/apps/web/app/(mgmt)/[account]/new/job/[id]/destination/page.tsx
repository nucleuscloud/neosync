'use client';
import { isValidConnectionPair } from '@/app/(mgmt)/[account]/connections/util';
import {
  getConnectionIdFromSource,
  getDestinationDetailsRecord,
  isDynamoDBConnection,
} from '@/app/(mgmt)/[account]/jobs/[id]/source/components/util';
import {
  getDefaultDestinationFormValueOptionsFromConnectionCase,
  toJobDestinationOptions,
} from '@/app/(mgmt)/[account]/jobs/util';
import PageHeader from '@/components/headers/PageHeader';
import DestinationOptionsForm from '@/components/jobs/Form/DestinationOptionsForm';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { PageProps } from '@/components/types';
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
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { splitConnections } from '@/libs/utils';
import { getErrorMessage } from '@/util/util';
import { NewDestinationFormValues } from '@/yup-validations/jobs';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  Connection,
  GetConnectionSchemaMapsResponse,
  JobDestination,
} from '@neosync/sdk';
import {
  createJobDestinationConnections,
  getConnections,
  getConnectionSchemaMaps,
  getJob,
} from '@neosync/sdk/connectquery';
import { Cross1Icon, PlusIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import * as Yup from 'yup';
import ConnectionSelectContent from '../../connect/ConnectionSelectContent';

const FormValues = Yup.object({
  destinations: Yup.array(NewDestinationFormValues).required(
    'Destinations are required.'
  ),
});
type FormValues = Yup.InferType<typeof FormValues>;

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { account } = useAccount();
  const router = useRouter();
  const { data, isLoading } = useQuery(getJob, { id }, { enabled: !!id });
  const { data: connectionsData, isLoading: isConnectionsLoading } = useQuery(
    getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const { mutateAsync: createJobConnections } = useMutation(
    createJobDestinationConnections
  );

  const connections = connectionsData?.connections ?? [];
  const connRecord = connections.reduce(
    (record, conn) => {
      record[conn.id] = conn;
      return record;
    },
    {} as Record<string, Connection>
  );
  const destinationConnectionIds = new Set(
    data?.job?.destinations.map((d) => d.connectionId)
  );
  const sourceConnectionId = getConnectionIdFromSource(data?.job?.source);
  const sourceConnection = sourceConnectionId
    ? connRecord[sourceConnectionId]
    : new Connection();

  const form = useForm({
    resolver: yupResolver<FormValues>(FormValues),
    defaultValues: {
      destinations: [{ connectionId: '', destinationOptions: {} }],
    },
  });

  const availableConnections = connections.filter(
    (c) =>
      c.id != sourceConnectionId &&
      !destinationConnectionIds?.has(c.id) &&
      isValidConnectionPair(sourceConnection, c)
  );

  const { append, remove } = useFieldArray({
    control: form.control,
    name: 'destinations',
  });

  const fields = form.watch('destinations');

  // Contains a list of the new destinations to be added that are specifically dynamo db connections
  const newDynamoDestConnections = fields
    .map((field) => connRecord[field.connectionId])
    .filter((conn) => !!conn && isDynamoDBConnection(conn));

  const { data: destinationConnectionSchemaMapsResp } = useQuery(
    getConnectionSchemaMaps,
    {
      requests: newDynamoDestConnections.map((conn) => ({
        connectionId: conn.id,
      })),
    },
    { enabled: newDynamoDestConnections.length > 0 }
  );

  async function onSubmit(values: FormValues): Promise<void> {
    try {
      const connMap = new Map(connections.map((c) => [c.id, c]));
      const job = await createJobConnections({
        jobId: id,
        destinations: values.destinations.map((d) => {
          return new JobDestination({
            connectionId: d.connectionId,
            options: toJobDestinationOptions(d, connMap.get(d.connectionId)),
          });
        }),
      });
      toast.success('Successfully created job destination(s)');
      if (job.job?.id) {
        router.push(`/${account?.name}/jobs/${job.job.id}/destinations`);
      } else {
        router.push(`/${account?.name}/jobs`);
      }
    } catch (err) {
      console.error(err);
      toast.error('Unable to create job destination(s)', {
        description: getErrorMessage(err),
      });
    }
  }

  const { postgres, mysql, s3, mongodb, gcpcs, dynamodb } =
    splitConnections(availableConnections);

  return (
    <div className="job-details-container mx-24">
      <div className="my-10">
        <PageHeader
          header="Create new Destination Connections"
          subHeadings={`Connect new destination datasources.`}
        />
      </div>

      {isLoading || isConnectionsLoading ? (
        <SkeletonTable />
      ) : (
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <div className="space-y-12">
              {fields.map((f, index) => {
                // not using the field here because it doesn't seem to always update when it needs to
                const connId = f.connectionId;
                const destOpts = f.destinationOptions;
                const destConnection = connRecord[connId] as
                  | Connection
                  | undefined;
                const destinationsErrors =
                  form.formState.errors.destinations ?? [];
                return (
                  <div key={index} className="space-y-4">
                    <div className="flex flex-row space-x-8">
                      <div className="basis-11/12">
                        <FormField
                          control={form.control}
                          name={`destinations.${index}.connectionId`}
                          render={({ field }) => (
                            <FormItem>
                              <FormControl>
                                <Select
                                  onValueChange={(value: string) => {
                                    field.onChange(value);
                                    const newDest = connRecord[value];
                                    const newOpts =
                                      getDefaultDestinationFormValueOptionsFromConnectionCase(
                                        newDest.connectionConfig?.config.case,
                                        () =>
                                          new Set(
                                            data?.job?.mappings.map(
                                              (mapping) => mapping.table
                                            )
                                          )
                                      );
                                    form.setValue(
                                      `destinations.${index}.destinationOptions`,
                                      newOpts,
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
                                      placeholder="Select a destination ..."
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
                                    />
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
                      </div>

                      <Button
                        variant="outline"
                        type="button"
                        onClick={() => {
                          remove(index);
                        }}
                      >
                        <Cross1Icon />
                      </Button>
                    </div>

                    <DestinationOptionsForm
                      connection={destConnection}
                      value={destOpts}
                      setValue={(newOpts) => {
                        form.setValue(
                          `destinations.${index}.destinationOptions`,
                          newOpts,
                          {
                            shouldDirty: true,
                            shouldTouch: true,
                            shouldValidate: true,
                          }
                        );
                      }}
                      hideDynamoDbTableMappings={
                        !isDynamoDBConnection(
                          destConnection ?? new Connection()
                        )
                      }
                      destinationDetailsRecord={getDestinationDetailsRecord(
                        fields.map((field) => ({
                          connectionId: field.connectionId,
                          id: field.connectionId,
                        })),
                        connRecord,
                        destinationConnectionSchemaMapsResp ??
                          new GetConnectionSchemaMapsResponse()
                      )}
                      errors={destinationsErrors[index]?.destinationOptions}
                    />
                  </div>
                );
              })}
              <div className="flex flex-row items-center justify-between w-full mt-4">
                <Button
                  variant="outline"
                  type="button"
                  onClick={() => {
                    append({ connectionId: '', destinationOptions: {} });
                  }}
                >
                  Add
                  <PlusIcon className="ml-2" />
                </Button>
                <Button
                  disabled={
                    !form.formState.isDirty ||
                    form.getValues().destinations.length === 0
                  }
                  type="submit"
                >
                  Save
                </Button>
              </div>
            </div>
          </form>
        </Form>
      )}
    </div>
  );
}
