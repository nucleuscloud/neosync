'use client';
import FormPersist from '@/app/(mgmt)/FormPersist';
import { getConnectionType } from '@/app/(mgmt)/[account]/connections/util';
import {
  getDefaultDestinationFormValueOptionsFromConnectionCase,
  getNewJobSessionKeys,
} from '@/app/(mgmt)/[account]/jobs/util';
import Spinner from '@/components/Spinner';
import TestConnectionBadge from '@/components/connections/TestConnectionBadge';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import DestinationOptionsForm from '@/components/jobs/Form/DestinationOptionsForm';
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
import { create } from '@bufbuild/protobuf';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckConnectionConfigByIdResponse,
  CheckConnectionConfigByIdResponseSchema,
  Code,
  ConnectError,
  ConnectionConfigSchema,
  ConnectionService,
} from '@neosync/sdk';
import { useRouter } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, use, useEffect, useState } from 'react';
import { Control, useForm, useWatch } from 'react-hook-form';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps, {
  getJobProgressSteps,
} from '../../../JobsProgressSteps';
import ConnectionSelectContent from '../../../connect/ConnectionSelectContent';
import { SingleTableConnectFormValues } from '../../../job-form-validations';

const NEW_CONNECTION_VALUE = 'new-connection';

export default function Page(props: PageProps): ReactElement {
  const searchParams = use(props.searchParams);
  const { account } = useAccount();
  const router = useRouter();
  const posthog = usePostHog();
  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/${account?.name}/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = getSingleOrUndefined(searchParams?.sessionId) ?? '';
  const sessionKeys = getNewJobSessionKeys(sessionPrefix);
  const formKey = sessionKeys.generate.connect;
  const [defaultValues] = useSessionStorage<SingleTableConnectFormValues>(
    formKey,
    {
      fkSourceConnectionId: '',
      destination: {
        connectionId: '',
        destinationOptions: {},
      },
    }
  );
  const [isSourceValidating, setIsSourceValidating] = useState<boolean>(false);

  const [sourceValidationResponse, setSourceValidationResponse] = useState<
    CheckConnectionConfigByIdResponse | undefined
  >();
  const [isDestinationValidating, setIsDestinationValidating] =
    useState<boolean>(false);

  const [destinationValidationResponse, setDestinationValidationResponse] =
    useState<CheckConnectionConfigByIdResponse | undefined>();

  const form = useForm({
    resolver: yupResolver<SingleTableConnectFormValues>(
      SingleTableConnectFormValues
    ),
    defaultValues,
  });

  const { isLoading: isConnectionsLoading, data: connectionsData } = useQuery(
    ConnectionService.method.getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const connections = connectionsData?.connections ?? [];

  async function onSubmit(_values: SingleTableConnectFormValues) {
    router.push(
      `/${account?.name}/new/job/generate/single/schema?sessionId=${sessionPrefix}`
    );
    posthog.capture('New Job Flow Connect Complete', { jobType: 'generate' });
  }

  const errors = form.formState.errors;

  const { mysql, postgres } = splitConnections(connections);

  const destOpts = form.watch('destination.destinationOptions');
  const shouldHideInitTableSchema = useShouldHideInitConnectionSchema(
    form.control
  );

  const { mutateAsync: checkConnectionConfig } = useMutation(
    ConnectionService.method.checkConnectionConfigById
  );

  return (
    <div
      id="newjobflowcontainer"
      className="px-12 md:px-24 lg:px-32 flex flex-col gap-5"
    >
      <FormPersist formKey={formKey} form={form} />
      <OverviewContainer
        Header={
          <PageHeader
            header="Connect"
            progressSteps={
              <JobsProgressSteps
                steps={getJobProgressSteps('generate-table')}
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
            <div className="space-y-0.5">
              <h2 className="text-xl font-semibold tracking-tight">
                Table Schema Connection
              </h2>
              <p className="text-muted-foreground text-sm">
                Choose a connection that will be used as a basis for the shape
                of data that is to be generated. This can be the same value as
                the destination.
              </p>
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
                        <div className="flex flex-row items-center gap-2">
                          <Select
                            onValueChange={async (value: string) => {
                              if (value === NEW_CONNECTION_VALUE) {
                                const destId = form.getValues(
                                  'destination.connectionId'
                                );

                                const urlParams = new URLSearchParams({
                                  returnTo: `/${account?.name}/new/job/generate/single/connect?sessionId=${sessionPrefix}&from=new-connection`,
                                });
                                const connection = connections.find(
                                  (c) => c.id === destId
                                );
                                const connType = getConnectionType(
                                  connection?.connectionConfig ??
                                    create(ConnectionConfigSchema, {})
                                );

                                if (connType) {
                                  urlParams.append('connectionType', connType);
                                }
                                if (
                                  urlParams.getAll('connectionType').length ===
                                  0
                                ) {
                                  urlParams.append(
                                    'connectionType',
                                    'pgConfig'
                                  );
                                  urlParams.append(
                                    'connectionType',
                                    'mysqlConfig'
                                  );
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
                                form.setValue(
                                  'destination.connectionId',
                                  value
                                );
                                const destConnection = connections.find(
                                  (c) => c.id === value
                                );
                                const destConnType = getConnectionType(
                                  destConnection?.connectionConfig ??
                                    create(ConnectionConfigSchema, {})
                                );
                                const newOpts =
                                  getDefaultDestinationFormValueOptionsFromConnectionCase(
                                    destConnType,
                                    () => new Set()
                                  );
                                form.setValue(
                                  'destination.destinationOptions',
                                  newOpts,
                                  {
                                    shouldDirty: true,
                                    shouldTouch: true,
                                    shouldValidate: true,
                                  }
                                );
                              }

                              setIsSourceValidating(true);
                              try {
                                const res = await checkConnectionConfig({
                                  id: form.getValues('fkSourceConnectionId'),
                                });
                                setSourceValidationResponse(res);
                              } catch (err) {
                                if (
                                  err instanceof ConnectError &&
                                  err.code === Code.InvalidArgument &&
                                  err.message.includes('unsupported operation')
                                ) {
                                  setSourceValidationResponse(undefined);
                                } else {
                                  setSourceValidationResponse(
                                    create(
                                      CheckConnectionConfigByIdResponseSchema,
                                      {
                                        isConnected: false,
                                        connectionError:
                                          err instanceof Error
                                            ? err.message
                                            : 'unknown error',
                                      }
                                    )
                                  );
                                }
                              } finally {
                                setIsSourceValidating(false);
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
                          <div className="relative pb-4">
                            {form.getValues('fkSourceConnectionId') &&
                              isSourceValidating && (
                                <Spinner className="text-black dark:text-white absolute" />
                              )}
                          </div>
                        </div>
                      )}
                    </FormControl>
                    <div className="flex">
                      {form.getValues('fkSourceConnectionId') &&
                        !isSourceValidating && (
                          <TestConnectionBadge
                            validationResponse={sourceValidationResponse}
                            connectionId={form.getValues(
                              'fkSourceConnectionId'
                            )}
                            accountName={account?.name ?? ''}
                          />
                        )}
                    </div>
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
                        <div className="flex flex-row items-center gap-2">
                          <Select
                            onValueChange={async (value: string) => {
                              if (value === NEW_CONNECTION_VALUE) {
                                const urlParams = new URLSearchParams({
                                  returnTo: `/${account?.name}/new/job/generate/single/connect?sessionId=${sessionPrefix}&from=new-connection`,
                                });
                                urlParams.append('connectionType', 'pgConfig');
                                urlParams.append(
                                  'connectionType',
                                  'mysqlConfig'
                                );
                                router.push(
                                  `/${account?.name}/new/connection?${urlParams.toString()}`
                                );
                                return;
                              }
                              field.onChange(value);
                              const destConnection = connections.find(
                                (c) => c.id === value
                              );
                              const destConnType = getConnectionType(
                                destConnection?.connectionConfig ??
                                  create(ConnectionConfigSchema, {})
                              );
                              const newOpts =
                                getDefaultDestinationFormValueOptionsFromConnectionCase(
                                  destConnType,
                                  () => new Set()
                                );
                              form.setValue(
                                'destination.destinationOptions',
                                newOpts,
                                {
                                  shouldDirty: true,
                                  shouldTouch: true,
                                  shouldValidate: true,
                                }
                              );
                              setIsDestinationValidating(true);
                              try {
                                const res = await checkConnectionConfig({
                                  id: form.getValues(
                                    'destination.connectionId'
                                  ),
                                });
                                setDestinationValidationResponse(res);
                              } catch (err) {
                                if (
                                  err instanceof ConnectError &&
                                  err.code === Code.InvalidArgument &&
                                  err.message.includes('unsupported operation')
                                ) {
                                  setDestinationValidationResponse(undefined);
                                } else {
                                  setDestinationValidationResponse(
                                    create(
                                      CheckConnectionConfigByIdResponseSchema,
                                      {
                                        isConnected: false,
                                        connectionError:
                                          err instanceof Error
                                            ? err.message
                                            : 'unknown error',
                                      }
                                    )
                                  );
                                }
                              } finally {
                                setIsDestinationValidating(false);
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
                          <div className="relative pb-4">
                            {form.getValues('destination.connectionId') &&
                              isDestinationValidating && (
                                <Spinner className="text-black dark:text-white absolute" />
                              )}
                          </div>
                        </div>
                      )}
                    </FormControl>
                    <div className="flex">
                      {form.getValues('destination.connectionId') &&
                        !isDestinationValidating && (
                          <TestConnectionBadge
                            validationResponse={
                              !destinationValidationResponse
                                ? sourceValidationResponse
                                : destinationValidationResponse
                            }
                            connectionId={form.getValues(
                              'destination.connectionId'
                            )}
                            accountName={account?.name ?? ''}
                          />
                        )}
                    </div>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <DestinationOptionsForm
                connection={connections.find(
                  (c) => c.id === form.getValues().destination.connectionId
                )}
                hideInitTableSchema={shouldHideInitTableSchema}
                value={destOpts}
                setValue={(newOpts) => {
                  form.setValue('destination.destinationOptions', newOpts, {
                    shouldDirty: true,
                    shouldTouch: true,
                    shouldValidate: true,
                  });
                }}
                hideDynamoDbTableMappings={true}
                destinationDetailsRecord={{}} // not used because we are hiding dynamodb table mappings
                errors={form.formState.errors?.destination?.destinationOptions}
              />
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
            <Button
              type="submit"
              disabled={(errors?.root?.message?.length ?? 0) > 0}
            >
              Next
            </Button>
          </div>
        </form>
      </Form>
    </div>
  );
}

function useShouldHideInitConnectionSchema(
  control: Control<SingleTableConnectFormValues>
): boolean {
  const [destinationConnectionid, fkSourceConnectionId] = useWatch({
    control,
    name: ['destination.connectionId', 'fkSourceConnectionId'],
  });
  return destinationConnectionid === fkSourceConnectionId;
}
