'use client';
import { getConnectionIdFromSource } from '@/app/(mgmt)/[account]/jobs/[id]/source/components/util';
import { toJobDestinationOptions } from '@/app/(mgmt)/[account]/jobs/util';
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
import { getErrorMessage } from '@/util/util';
import { NewDestinationFormValues } from '@/yup-validations/jobs';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { Connection, JobDestination } from '@neosync/sdk';
import {
  createJobDestinationConnections,
  getConnections,
  getJob,
} from '@neosync/sdk/connectquery';
import { Cross1Icon, PlusIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement, useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import * as Yup from 'yup';

const FormValues = Yup.object({
  destinations: Yup.array(NewDestinationFormValues).required(),
});
type FormValues = Yup.InferType<typeof FormValues>;

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { account } = useAccount();
  const { toast } = useToast();
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

  const [currConnection, setCurrConnection] = useState<
    Connection | undefined
  >();

  const connections = connectionsData?.connections ?? [];
  const destinationIds = new Set(
    data?.job?.destinations.map((d) => d.connectionId)
  );
  const sourceConnectionId = getConnectionIdFromSource(data?.job?.source);
  const form = useForm({
    resolver: yupResolver<FormValues>(FormValues),
    defaultValues: {
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
      const job = await createJobConnections({
        jobId: id,
        destinations: values.destinations.map((d) => {
          return new JobDestination({
            connectionId: d.connectionId,
            options: toJobDestinationOptions(
              d,
              connections.find((c) => c.id === d.connectionId)
            ),
          });
        }),
      });
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
          subHeadings={`Connect new destination datasources.`}
        />
      </div>

      {isLoading || isConnectionsLoading ? (
        <SkeletonTable />
      ) : (
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <div className="space-y-12">
              {fields.map((_, index) => {
                const destOpts = form.watch(
                  `destinations.${index}.destinationOptions`
                );
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
                                      {}
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
                      connection={currConnection}
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
