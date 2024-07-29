'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import SourceOptionsForm from '@/components/jobs/Form/SourceOptionsForm';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
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
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { getSingleOrUndefined, splitConnections } from '@/libs/utils';
import { useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { ConnectionConfig } from '@neosync/sdk';
import { getConnections } from '@neosync/sdk/connectquery';
import { Cross2Icon, PlusIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import DestinationOptionsForm from '../../../../../../components/jobs/Form/DestinationOptionsForm';
import {
  DESTINATION_ONLY_CONNECTION_TYPES,
  getConnectionType,
} from '../../../connections/util';
import { getNewJobSessionKeys } from '../../../jobs/util';
import JobsProgressSteps, { getJobProgressSteps } from '../JobsProgressSteps';
import { ConnectFormValues } from '../schema';
import ConnectionSelectContent from './ConnectionSelectContent';

const NEW_CONNECTION_VALUE = 'new-connection';

const isBrowser = () => typeof window !== 'undefined';

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/${account?.name}/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = getSingleOrUndefined(searchParams?.sessionId) ?? '';
  const sessionKeys = getNewJobSessionKeys(sessionPrefix);
  const sessionKey = sessionKeys.dataSync.connect;
  const [defaultValues] = useSessionStorage<ConnectFormValues>(sessionKey, {
    sourceId: '',
    sourceOptions: {},
    destinations: [{ connectionId: '', destinationOptions: {} }],
  });

  const { isLoading: isConnectionsLoading, data: connectionsData } = useQuery(
    getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const connections = connectionsData?.connections ?? [];

  const form = useForm<ConnectFormValues>({
    mode: 'onChange',
    resolver: yupResolver<ConnectFormValues>(ConnectFormValues),
    values: defaultValues,
    context: { connections },
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: 'destinations',
  });

  useFormPersist(sessionKey, {
    watch: form.watch,
    setValue: form.setValue,
    storage: isBrowser() ? window.sessionStorage : undefined,
  });
  const posthog = usePostHog();

  async function onSubmit(_values: ConnectFormValues) {
    posthog.capture('New Job Flow Connect Complete', { jobType: 'data-sync' });
    router.push(`/${account?.name}/new/job/schema?sessionId=${sessionPrefix}`);
  }

  const { postgres, mysql, s3, mongodb, gcpcs, dynamodb } =
    splitConnections(connections);

  return (
    <div
      id="newjobflowcontainer"
      className="px-12 md:px-24 lg:px-48 xl:px-64 flex flex-col gap-5"
    >
      <OverviewContainer
        Header={
          <PageHeader
            header="Connect"
            progressSteps={
              <JobsProgressSteps
                steps={getJobProgressSteps('data-sync', false)}
                stepName={'connect'}
              />
            }
          />
        }
        containerClassName="connect-page"
      >
        <div />
      </OverviewContainer>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <div
            className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4`}
          >
            <div>
              <div>
                <div className="space-y-0.5">
                  <h2 className="text-xl font-semibold tracking-tight">
                    Source
                  </h2>
                  <p className="text-muted-foreground text-sm">
                    The location of the source data set.
                  </p>
                </div>
              </div>
            </div>
            <div className="space-y-4 col-span-2">
              <FormField
                control={form.control}
                name="sourceId"
                render={({ field }) => (
                  <FormItem>
                    <FormControl>
                      {isConnectionsLoading ? (
                        <Skeleton className="w-full h-12 rounded-lg" />
                      ) : (
                        <Select
                          name={field.name}
                          disabled={field.disabled}
                          onValueChange={(value: string) => {
                            if (!value) {
                              return;
                            }
                            if (value === NEW_CONNECTION_VALUE) {
                              const destIds = new Set(
                                form
                                  .getValues('destinations')
                                  .map((d) => d.connectionId)
                              );

                              const urlParams = new URLSearchParams({
                                returnTo: `/${account?.name}/new/job/connect?sessionId=${sessionPrefix}&from=new-connection`,
                              });

                              const connTypes = new Set(
                                connections
                                  .filter((c) => destIds.has(c.id))
                                  .map((c) =>
                                    getConnectionType(
                                      c.connectionConfig ??
                                        new ConnectionConfig()
                                    )
                                  )
                              );
                              connTypes.forEach((connType) => {
                                if (
                                  connType &&
                                  !DESTINATION_ONLY_CONNECTION_TYPES.has(
                                    connType
                                  )
                                ) {
                                  urlParams.append('connectionType', connType);
                                }
                              });
                              if (
                                urlParams.getAll('connectionType').length === 0
                              ) {
                                urlParams.append('connectionType', 'postgres');
                                urlParams.append('connectionType', 'mysql');
                                urlParams.append('connectionType', 'mongodb');
                                urlParams.append('connectionType', 'dynamodb');
                              }

                              router.push(
                                `/${account?.name}/new/connection?${urlParams.toString()}`
                              );
                              return;
                            }
                            field.onChange(value);
                            form.setValue('sourceOptions', {
                              haltOnNewColumnAddition: false,
                            });
                          }}
                          value={field.value}
                        >
                          <SelectTrigger>
                            <SelectValue
                              ref={field.ref}
                              placeholder="Select a source ..."
                            />
                          </SelectTrigger>
                          <SelectContent>
                            <ConnectionSelectContent
                              postgres={postgres}
                              mysql={mysql}
                              mongodb={mongodb}
                              dynamodb={dynamodb}
                              newConnectionValue={NEW_CONNECTION_VALUE}
                            />
                          </SelectContent>
                        </Select>
                      )}
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <SourceOptionsForm
                connection={connections.find(
                  (c) => c.id === form.getValues().sourceId
                )}
              />
            </div>
          </div>
          <Separator className="my-6" />

          <div
            className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4`}
          >
            <div className="space-y-0.5">
              <h2 className="text-xl font-semibold tracking-tight">
                Destination(s)
              </h2>
              <p className="text-muted-foreground text-sm">
                Where the data set should be synced.
              </p>
            </div>
            <div className="space-y-12 col-span-2">
              {fields.map((val, index) => {
                const destOpts = form.watch(
                  `destinations.${index}.destinationOptions`
                );
                return (
                  <div className="space-y-4 col-span-2" key={val.id}>
                    <div className="flex flew-row space-x-4">
                      <div className="basis-11/12">
                        <FormField
                          control={form.control}
                          name={`destinations.${index}.connectionId`}
                          render={({ field }) => (
                            <FormItem>
                              <FormControl>
                                {isConnectionsLoading ? (
                                  <Skeleton className="w-full h-12 rounded-lg" />
                                ) : (
                                  <Select
                                    name={field.name}
                                    disabled={field.disabled}
                                    onValueChange={(value: string) => {
                                      if (!value) {
                                        return;
                                      }
                                      const sourceId =
                                        form.getValues('sourceId');
                                      const connection = connections.find(
                                        (c) => c.id === sourceId
                                      );
                                      const connType = getConnectionType(
                                        connection?.connectionConfig ??
                                          new ConnectionConfig()
                                      );

                                      if (value === NEW_CONNECTION_VALUE) {
                                        const urlParams = new URLSearchParams({
                                          returnTo: `/${account?.name}/new/job/connect?sessionId=${sessionPrefix}&from=new-connection`,
                                        });

                                        if (connType) {
                                          urlParams.append(
                                            'connectionType',
                                            connType
                                          );
                                        }
                                        if (
                                          urlParams.getAll('connectionType')
                                            .length === 0
                                        ) {
                                          urlParams.append(
                                            'connectionType',
                                            'postgres'
                                          );
                                          urlParams.append(
                                            'connectionType',
                                            'mysql'
                                          );
                                          urlParams.append(
                                            'connectionType',
                                            'aws-s3'
                                          );
                                          urlParams.append(
                                            'connectionType',
                                            'gcp-cloud-storage'
                                          );
                                          urlParams.append(
                                            'connectionType',
                                            'mongodb'
                                          );
                                          urlParams.append(
                                            'connectionType',
                                            'dynamodb'
                                          );
                                        }

                                        router.push(
                                          `/${account?.name}/new/connection?${urlParams.toString()}`
                                        );
                                        return;
                                      }
                                      // set values
                                      field.onChange(value);
                                      const destConnection = connections.find(
                                        (c) => c.id === value
                                      );
                                      const destConnType = getConnectionType(
                                        destConnection?.connectionConfig ??
                                          new ConnectionConfig()
                                      );
                                      if (destConnType === 'postgres') {
                                        form.setValue(
                                          `destinations.${index}.destinationOptions`,
                                          {
                                            postgres: {
                                              truncateBeforeInsert: false,
                                              truncateCascade: false,
                                              initTableSchema: false,
                                              onConflictDoNothing: false,
                                            },
                                          },
                                          {
                                            shouldDirty: true,
                                            shouldTouch: true,
                                            shouldValidate: true,
                                          }
                                        );
                                      } else if (destConnType === 'mysql') {
                                        form.setValue(
                                          `destinations.${index}.destinationOptions`,
                                          {
                                            mysql: {
                                              truncateBeforeInsert: false,
                                              initTableSchema: false,
                                              onConflictDoNothing: false,
                                            },
                                          },
                                          {
                                            shouldDirty: true,
                                            shouldTouch: true,
                                            shouldValidate: true,
                                          }
                                        );
                                      } else if (destConnType === 'dynamodb') {
                                        form.setValue(
                                          `destinations.${index}.destinationOptions`,
                                          {
                                            dynamodb: {
                                              tableMappings: [],
                                            },
                                          },
                                          {
                                            shouldDirty: true,
                                            shouldTouch: true,
                                            shouldValidate: true,
                                          }
                                        );
                                      }
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
                                        newConnectionValue={
                                          NEW_CONNECTION_VALUE
                                        }
                                      />
                                    </SelectContent>
                                  </Select>
                                )}
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                      </div>
                      <div>
                        <Button
                          type="button"
                          variant="outline"
                          disabled={fields.length === 1}
                          onClick={() => {
                            if (fields.length != 1) {
                              remove(index);
                            }
                          }}
                        >
                          <Cross2Icon className="w-4 h-4" />
                        </Button>
                      </div>
                    </div>
                    <DestinationOptionsForm
                      connection={connections.find(
                        (c) =>
                          c.id ==
                          form.getValues().destinations[index].connectionId
                      )}
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
                      hideDynamoDbTableMappings={true}
                      destinationDetailsRecord={{}} // not used beacause we are hiding dynamodb table mappings
                    />
                  </div>
                );
              })}

              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  append({
                    connectionId: '',
                    destinationOptions: {},
                  });
                }}
              >
                Add
                <PlusIcon className="ml-2 w-4 h-4" />
              </Button>
            </div>
          </div>

          <div className="flex flex-row gap-1 justify-between">
            <Button
              type="button"
              onClick={() => {
                if (searchParams?.from === 'new-connection') {
                  router.push(
                    `/${account?.name}/new/job/define?sessionId=${sessionPrefix}`
                  );
                  return;
                }
                router.back();
              }}
            >
              Back
            </Button>
            <Button type="submit" disabled={!form.formState.isValid}>
              Next
            </Button>
          </div>
        </form>
      </Form>
    </div>
  );
}
