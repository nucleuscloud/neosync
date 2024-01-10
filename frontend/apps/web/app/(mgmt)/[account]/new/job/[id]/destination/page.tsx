'use client';
import { getConnectionIdFromSource } from '@/app/(mgmt)/[account]/jobs/[id]/source/components/util';
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
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useToast } from '@/components/ui/use-toast';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { getErrorMessage } from '@/util/util';
import {
  DESTINATION_FORM_SCHEMA,
  toJobDestinationOptions,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  Connection,
  CreateJobDestinationConnectionsRequest,
  CreateJobDestinationConnectionsResponse,
  JobDestination,
} from '@neosync/sdk';
import { Cross1Icon, PlusIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement, useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import * as Yup from 'yup';

const FORM_SCHEMA = Yup.object({
  jobId: Yup.string().required(),
  destinations: Yup.array(DESTINATION_FORM_SCHEMA).required(),
});
type FormValues = Yup.InferType<typeof FORM_SCHEMA>;

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { account } = useAccount();
  const { toast } = useToast();
  const router = useRouter();
  const { data, isLoading } = useGetJob(account?.id ?? '', id);
  const { isLoading: isConnectionsLoading, data: connectionsData } =
    useGetConnections(account?.id ?? '');

  const [currConnection, setCurrConnection] = useState<
    Connection | undefined
  >();

  const connections = connectionsData?.connections ?? [];
  const destinationIds = new Set(
    data?.job?.destinations.map((d) => d.connectionId)
  );
  const sourceConnectionId = getConnectionIdFromSource(data?.job?.source);
  const form = useForm({
    resolver: yupResolver<FormValues>(FORM_SCHEMA),
    defaultValues: {
      jobId: id,
      destinations: [{ connectionId: '', destinationOptions: {} }],
    },
  });

  const availableConnections = connections.filter(
    (c) => c.id != sourceConnectionId && !destinationIds?.has(c.id)
  );

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: 'destinations',
  });

  async function onSubmit(values: FormValues) {
    try {
      const job = await createJobConnections(
        id,
        values,
        connections,
        account?.id ?? ''
      );
      if (job.job?.id) {
        router.push(`/${account?.name}/jobs/${job.job.id}/destinations`);
      } else {
        router.push(`/${account?.name}/jobs`);
      }
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to create job destinations',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <div className="job-details-container mx-24">
      <div className="my-10">
        <PageHeader
          header="Create new Destination Connections"
          description={`Connect new destination datasources.`}
        />
      </div>

      {isLoading || isConnectionsLoading ? (
        <SkeletonTable />
      ) : (
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <div className="space-y-12">
              {fields.map((_, index) => {
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
                                    setCurrConnection(
                                      connections.find(
                                        (c) =>
                                          c.id ==
                                          form.getValues().destinations[index]
                                            .connectionId
                                      )
                                    );
                                    form.setValue(
                                      `destinations.${index}.destinationOptions`,
                                      {
                                        truncateBeforeInsert: false,
                                        truncateCascade: false,
                                        initTableSchema: false,
                                      }
                                    );
                                  }}
                                  value={field.value}
                                >
                                  <SelectTrigger>
                                    <SelectValue placeholder="Select a destination ..." />
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
                      index={index}
                      connection={currConnection}
                      maxColNum={3}
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
                    form.getValues().destinations.length == 0
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

async function createJobConnections(
  jobId: string,
  values: FormValues,
  connections: Connection[],
  accountId: string
): Promise<CreateJobDestinationConnectionsResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/jobs/${jobId}/destination-connections`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(
        new CreateJobDestinationConnectionsRequest({
          jobId: jobId,
          destinations: values.destinations.map((d) => {
            return new JobDestination({
              connectionId: d.connectionId,
              options: toJobDestinationOptions(
                d,
                connections.find((c) => c.id == d.connectionId)
              ),
            });
          }),
        })
      ),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CreateJobDestinationConnectionsResponse.fromJson(await res.json());
}
