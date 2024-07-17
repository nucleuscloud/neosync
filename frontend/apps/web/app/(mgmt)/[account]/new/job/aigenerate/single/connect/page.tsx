'use client';
import { getConnectionType } from '@/app/(mgmt)/[account]/connections/util';
import { getNewJobSessionKeys } from '@/app/(mgmt)/[account]/jobs/util';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import DestinationOptionsForm from '@/components/jobs/Form/DestinationOptionsForm';
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
import { useRouter } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect } from 'react';
import { Control, useForm, useWatch } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps, {
  getJobProgressSteps,
} from '../../../JobsProgressSteps';
import ConnectionSelectContent from '../../../connect/ConnectionSelectContent';
import { SingleTableAiConnectFormValues } from '../../../schema';

const NEW_CONNECTION_VALUE = 'new-connection';

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/${account?.name}/new/job`);
    }
  }, [searchParams?.sessionId]);
  const posthog = usePostHog();

  const sessionPrefix = getSingleOrUndefined(searchParams?.sessionId) ?? '';
  const formKey = getNewJobSessionKeys(sessionPrefix).aigenerate.connect;
  const [defaultValues] = useSessionStorage<SingleTableAiConnectFormValues>(
    formKey,
    {
      sourceId: '',
      fkSourceConnectionId: '',
      destination: {
        connectionId: '',
        destinationOptions: {},
      },
    }
  );

  const form = useForm<SingleTableAiConnectFormValues>({
    resolver: yupResolver<SingleTableAiConnectFormValues>(
      SingleTableAiConnectFormValues
    ),
    defaultValues,
  });

  useFormPersist(formKey, {
    watch: form.watch,
    setValue: form.setValue,
    storage: window.sessionStorage,
  });
  const { isLoading: isConnectionsLoading, data: connectionsData } = useQuery(
    getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const connections = connectionsData?.connections ?? [];

  function onSubmit(_values: SingleTableAiConnectFormValues) {
    router.push(
      `/${account?.name}/new/job/aigenerate/single/schema?sessionId=${sessionPrefix}`
    );
    posthog.capture('New Job Flow Connect Complete', {
      jobType: 'ai-generate',
    });
  }

  const { mysql, postgres, openai } = splitConnections(connections);

  const destOpts = form.watch('destination.destinationOptions');
  const shouldHideInitTableSchema = useShouldHideInitConnectionSchema(
    form.control
  );

  return (
    <div
      id="newjobflowcontainer"
      className="px-12 md:px-24 lg:px-32 flex flex-col gap-5"
    >
      <OverviewContainer
        Header={
          <PageHeader
            header="Connect"
            progressSteps={
              <JobsProgressSteps
                steps={getJobProgressSteps('ai-generate-table')}
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
                    Choose an OpenAI Compatible SDK to generate data from.
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
                              const urlParams = new URLSearchParams({
                                returnTo: `/${account?.name}/new/job/aigenerate/single/connect?sessionId=${sessionPrefix}&from=new-connection`,
                                connectionType: 'openai',
                              });
                              router.push(
                                `/${account?.name}/new/connection?${urlParams.toString()}`
                              );
                              return;
                            }
                            field.onChange(value);
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
                              openai={openai}
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
            <div>
              <div>
                <div className="space-y-0.5">
                  <h2 className="text-xl font-semibold tracking-tight">
                    Table Schema Connection
                  </h2>
                  <p className="text-muted-foreground text-sm">
                    Choose a connection that will be used as a basis for the
                    shape of data that is to be generated. This can be the same
                    value as the destination.
                  </p>
                </div>
              </div>
            </div>
            <div className="space-y-4 col-span-2">
              <FormField
                control={form.control}
                name="fkSourceConnectionId"
                render={({ field }) => (
                  <FormItem>
                    <FormControl>
                      {isConnectionsLoading ? (
                        <Skeleton className="w-full h-12 rounded-lg" />
                      ) : (
                        <Select
                          onValueChange={(value: string) => {
                            if (value === NEW_CONNECTION_VALUE) {
                              const destId = form.getValues(
                                'destination.connectionId'
                              );

                              const urlParams = new URLSearchParams({
                                returnTo: `/${account?.name}/new/job/aigenerate/single/connect?sessionId=${sessionPrefix}&from=new-connection`,
                              });
                              const connection = connections.find(
                                (c) => c.id === destId
                              );
                              const connType = getConnectionType(
                                connection?.connectionConfig ??
                                  new ConnectionConfig()
                              );
                              if (connType) {
                                urlParams.append('connectionType', connType);
                              }
                              if (
                                urlParams.getAll('connectionType').length === 0
                              ) {
                                urlParams.append('connectionType', 'postgres');
                                urlParams.append('connectionType', 'mysql');
                              }
                              router.push(
                                `/${account?.name}/new/connection?${urlParams.toString()}`
                              );
                              return;
                            }
                            field.onChange(value);
                            const destId = form.getValues(
                              'destination.connectionId'
                            );
                            if (!destId) {
                              form.setValue('destination.connectionId', value);
                            }
                          }}
                          value={field.value}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Select a connection ..." />
                          </SelectTrigger>
                          <SelectContent>
                            <ConnectionSelectContent
                              postgres={postgres}
                              mysql={mysql}
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
            </div>
          </div>
          <Separator className="my-6" />

          <div
            className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4`}
          >
            <div>
              <div>
                <div className="space-y-0.5">
                  <h2 className="text-xl font-semibold tracking-tight">
                    Destination
                  </h2>
                  <p className="text-muted-foreground text-sm">
                    The location the generated data will be synced to.
                  </p>
                </div>
              </div>
            </div>
            <div className="space-y-4 col-span-2">
              <FormField
                control={form.control}
                name="destination.connectionId"
                render={({ field }) => (
                  <FormItem>
                    <FormControl>
                      {isConnectionsLoading ? (
                        <Skeleton className="w-full h-12 rounded-lg" />
                      ) : (
                        <Select
                          onValueChange={(value: string) => {
                            if (value === NEW_CONNECTION_VALUE) {
                              const fkSourceConnId = form.getValues(
                                'fkSourceConnectionId'
                              );

                              const urlParams = new URLSearchParams({
                                returnTo: `/${account?.name}/new/job/aigenerate/single/connect?sessionId=${sessionPrefix}&from=new-connection`,
                              });
                              const connection = connections.find(
                                (c) => c.id === fkSourceConnId
                              );
                              const connType = getConnectionType(
                                connection?.connectionConfig ??
                                  new ConnectionConfig()
                              );
                              if (connType) {
                                urlParams.append('connectionType', connType);
                              }
                              if (
                                urlParams.getAll('connectionType').length === 0
                              ) {
                                urlParams.append('connectionType', 'postgres');
                                urlParams.append('connectionType', 'mysql');
                              }
                              router.push(
                                `/${account?.name}/new/connection?${urlParams.toString()}`
                              );
                              return;
                            }
                            field.onChange(value);
                            form.setValue('destination.destinationOptions', {
                              initTableSchema: false,
                              truncateBeforeInsert: false,
                              truncateCascade: false,
                              onConflictDoNothing: false,
                            });
                          }}
                          value={field.value}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Select a connection ..." />
                          </SelectTrigger>
                          <SelectContent>
                            <ConnectionSelectContent
                              postgres={postgres}
                              mysql={mysql}
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

              <DestinationOptionsForm
                connection={connections.find(
                  (c) => c.id === form.getValues().destination.connectionId
                )}
                hideInitTableSchema={shouldHideInitTableSchema}
                value={{
                  initTableSchema: destOpts.initTableSchema ?? false,
                  onConflictDoNothing: destOpts.onConflictDoNothing ?? false,
                  truncateBeforeInsert: destOpts.truncateBeforeInsert ?? false,
                  truncateCascade: destOpts.truncateCascade ?? false,
                }}
                setValue={(newOpts) => {
                  form.setValue(
                    'destination.destinationOptions',
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
              />
            </div>
          </div>
          <Separator className="my-6" />

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
            <Button type="submit">Next</Button>
          </div>
        </form>
      </Form>
    </div>
  );
}

function useShouldHideInitConnectionSchema(
  control: Control<SingleTableAiConnectFormValues>
): boolean {
  const [destinationConnectionid, fkSourceConnectionId] = useWatch({
    control,
    name: ['destination.connectionId', 'fkSourceConnectionId'],
  });
  return destinationConnectionid === fkSourceConnectionId;
}
