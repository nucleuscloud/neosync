'use client';
import FormPersist from '@/app/(mgmt)/FormPersist';
import Spinner from '@/components/Spinner';
import TestConnectionBadge from '@/components/connections/TestConnectionBadge';
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
import { cn, getSingleOrUndefined, splitConnections } from '@/libs/utils';
import { create } from '@bufbuild/protobuf';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckConnectionConfigByIdResponse,
  CheckConnectionConfigByIdResponseSchema,
  Code,
  ConnectError,
  ConnectionConfigSchema,
  ConnectionSchema,
  ConnectionService,
} from '@neosync/sdk';
import { Cross2Icon, PlusIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import { useSessionStorage } from 'usehooks-ts';
import DestinationOptionsForm from '../../../../../../components/jobs/Form/DestinationOptionsForm';
import {
  getAllowedSyncDestinationTypes,
  getAllowedSyncSourceTypes,
  getConnectionType,
} from '../../../connections/util';
import {
  getDefaultDestinationFormValueOptionsFromConnectionCase,
  getDefaultUnmappedTransformConfig,
  getNewJobSessionKeys,
} from '../../../jobs/util';
import JobsProgressSteps, { getJobProgressSteps } from '../JobsProgressSteps';
import { ConnectFormValues } from '../job-form-validations';
import ConnectionSelectContent from './ConnectionSelectContent';

const NEW_CONNECTION_VALUE = 'new-connection';

interface DestinationValidationState {
  isValidating: boolean;
  response?: CheckConnectionConfigByIdResponse;
}

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
  const [isSourceValidating, setIsSourceValidating] = useState<boolean>(false);

  const [sourceValidationResponse, setSourceValidationResponse] = useState<
    CheckConnectionConfigByIdResponse | undefined
  >();

  const [destinationValidation, setDestinationValidation] = useState<
    Record<string, DestinationValidationState>
  >({});

  const { isLoading: isConnectionsLoading, data: connectionsData } = useQuery(
    ConnectionService.method.getConnections,
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

  const posthog = usePostHog();

  async function onSubmit(_values: ConnectFormValues) {
    posthog.capture('New Job Flow Connect Complete', { jobType: 'data-sync' });
    router.push(`/${account?.name}/new/job/schema?sessionId=${sessionPrefix}`);
  }

  const { postgres, mysql, s3, mongodb, gcpcs, dynamodb, mssql } =
    splitConnections(connections);

  const { mutateAsync: checkConnectionConfig } = useMutation(
    ConnectionService.method.checkConnectionConfigById
  );

  return (
    <div
      id="newjobflowcontainer"
      className="px-12 md:px-24 lg:px-48 xl:px-64 flex flex-col gap-5"
    >
      <FormPersist formKey={sessionKey} form={form} />
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
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <div
            className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4`}
          >
            <div className="space-y-0.5">
              <h2 className="text-xl font-semibold tracking-tight">Source</h2>
              <p className="text-muted-foreground text-sm">
                The location of the source data.
              </p>
            </div>
            <div className=" col-span-2">
              <FormField
                control={form.control}
                name="sourceId"
                render={({ field }) => (
                  <FormItem>
                    <FormControl>
                      {isConnectionsLoading ? (
                        <Skeleton className="w-full h-12 rounded-lg" />
                      ) : (
                        <div className="flex flex-row items-center gap-2">
                          <Select
                            name={field.name}
                            disabled={field.disabled}
                            onValueChange={async (value: string) => {
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
                                          create(ConnectionConfigSchema, {})
                                      )
                                    )
                                    .filter((x) => !!x)
                                );
                                const allowedSourceTypes =
                                  getAllowedSyncSourceTypes(
                                    Array.from(connTypes)
                                  );
                                allowedSourceTypes.forEach((sourceType) =>
                                  urlParams.append('connectionType', sourceType)
                                );

                                router.push(
                                  `/${account?.name}/new/connection?${urlParams.toString()}`
                                );
                                return;
                              }
                              field.onChange(value);
                              const connection =
                                connections.find((c) => c.id === value) ??
                                create(ConnectionSchema, {});
                              const connectionType = getConnectionType(
                                connection.connectionConfig ??
                                  create(ConnectionConfigSchema, {})
                              );
                              setIsSourceValidating(true);
                              try {
                                const res = await checkConnectionConfig({
                                  id: form.getValues('sourceId'),
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

                              if (connectionType === 'pgConfig') {
                                form.setValue(
                                  'sourceOptions',
                                  {
                                    postgres: {
                                      newColumnAdditionStrategy: 'halt',
                                    },
                                  },
                                  {
                                    shouldDirty: true,
                                    shouldTouch: true,
                                    shouldValidate: true,
                                  }
                                );
                              } else if (connectionType === 'mysqlConfig') {
                                form.setValue(
                                  'sourceOptions',
                                  {
                                    mysql: {
                                      haltOnNewColumnAddition: false,
                                    },
                                  },
                                  {
                                    shouldDirty: true,
                                    shouldTouch: true,
                                    shouldValidate: true,
                                  }
                                );
                              } else if (connectionType === 'dynamodbConfig') {
                                form.setValue(
                                  'sourceOptions',
                                  {
                                    dynamodb: {
                                      unmappedTransformConfig:
                                        getDefaultUnmappedTransformConfig(),
                                      enableConsistentRead: false,
                                    },
                                  },
                                  {
                                    shouldDirty: true,
                                    shouldTouch: true,
                                    shouldValidate: true,
                                  }
                                );
                              } else if (connectionType === 'mssqlConfig') {
                                form.setValue(
                                  'sourceOptions',
                                  {
                                    mssql: {
                                      haltOnNewColumnAddition: false,
                                    },
                                  },
                                  {
                                    shouldDirty: true,
                                    shouldTouch: true,
                                    shouldValidate: true,
                                  }
                                );
                              } else {
                                form.setValue(
                                  'sourceOptions',
                                  {},
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
                            <SelectTrigger
                              className={cn(
                                field.value
                                  ? undefined
                                  : 'text-muted-foreground'
                              )}
                            >
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
                                mssql={mssql}
                                newConnectionValue={NEW_CONNECTION_VALUE}
                              />
                            </SelectContent>
                          </Select>
                          <div className="relative pb-4">
                            {form.getValues('sourceId') &&
                              isSourceValidating && (
                                <Spinner className="text-black dark:text-white absolute" />
                              )}
                          </div>
                        </div>
                      )}
                    </FormControl>
                    <div className="flex">
                      {form.getValues('sourceId') && !isSourceValidating && (
                        <TestConnectionBadge
                          validationResponse={sourceValidationResponse}
                          connectionId={form.getValues('sourceId')}
                          accountName={account?.name ?? ''}
                        />
                      )}
                    </div>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="sourceOptions"
                render={({ field }) => (
                  <FormItem>
                    <FormControl>
                      <SourceOptionsForm
                        connection={connections.find(
                          (c) => c.id === form.getValues().sourceId
                        )}
                        value={field.value}
                        setValue={(newOpts) => field.onChange(newOpts)}
                      />
                    </FormControl>
                  </FormItem>
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
                Where the data should be synced.
              </p>
            </div>
            <div className="space-y-12 col-span-2">
              {fields.map((val, index) => {
                const destErrs = form.formState.errors?.destinations ?? [];
                return (
                  <div className="space-y-4 col-span-2" key={val.id}>
                    <div className="flex flex-row gap-2">
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
                                  <div className="flex flex-row items-center gap-2">
                                    <Select
                                      name={field.name}
                                      disabled={field.disabled}
                                      onValueChange={async (value: string) => {
                                        if (!value) {
                                          return;
                                        }
                                        const sourceId =
                                          form.getValues('sourceId');
                                        const connection = connections.find(
                                          (c) => c.id === sourceId
                                        );

                                        if (value === NEW_CONNECTION_VALUE) {
                                          const urlParams = new URLSearchParams(
                                            {
                                              returnTo: `/${account?.name}/new/job/connect?sessionId=${sessionPrefix}&from=new-connection`,
                                            }
                                          );

                                          const allowedDestinationConnectionTypes =
                                            getAllowedSyncDestinationTypes(
                                              connection?.connectionConfig
                                                ?.config.case
                                            );
                                          allowedDestinationConnectionTypes.forEach(
                                            (ct) =>
                                              urlParams.append(
                                                'connectionType',
                                                ct
                                              )
                                          );

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
                                            create(ConnectionConfigSchema, {})
                                        );

                                        setDestinationValidation(
                                          (prevState) => ({
                                            ...prevState,
                                            [value]: {
                                              isValidating: true,
                                              response: create(
                                                CheckConnectionConfigByIdResponseSchema,
                                                {}
                                              ),
                                            },
                                          })
                                        );
                                        try {
                                          const res =
                                            await checkConnectionConfig({
                                              id: value,
                                            });
                                          setDestinationValidation(
                                            (prevState) => ({
                                              ...prevState,
                                              [value]: {
                                                isValidating: false,
                                                response: res,
                                              },
                                            })
                                          );
                                        } catch (err) {
                                          if (
                                            err instanceof ConnectError &&
                                            err.code === Code.InvalidArgument &&
                                            err.message.includes(
                                              'unsupported operation'
                                            )
                                          ) {
                                            setDestinationValidation(
                                              (prevState) => ({
                                                ...prevState,
                                                [value]: {
                                                  isValidating: false,
                                                  response: undefined,
                                                },
                                              })
                                            );
                                          } else {
                                            setDestinationValidation(
                                              (prevState) => ({
                                                ...prevState,
                                                [value]: {
                                                  isValidating: false,
                                                  response: create(
                                                    CheckConnectionConfigByIdResponseSchema,
                                                    {
                                                      isConnected: false,
                                                      connectionError:
                                                        err instanceof Error
                                                          ? err.message
                                                          : 'unknown error',
                                                    }
                                                  ),
                                                },
                                              })
                                            );
                                          }
                                        } finally {
                                          setIsSourceValidating(false);
                                        }

                                        const newOpts =
                                          getDefaultDestinationFormValueOptionsFromConnectionCase(
                                            destConnType,
                                            () => new Set()
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
                                      <SelectTrigger
                                        className={cn(
                                          field.value
                                            ? undefined
                                            : 'text-muted-foreground'
                                        )}
                                      >
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
                                          mssql={mssql}
                                          newConnectionValue={
                                            NEW_CONNECTION_VALUE
                                          }
                                        />
                                      </SelectContent>
                                    </Select>
                                    <div className="relative pb-4">
                                      {form.getValues(
                                        `destinations.${index}.connectionId`
                                      ) &&
                                        destinationValidation[
                                          form.getValues(
                                            `destinations.${index}.connectionId`
                                          )
                                        ]?.isValidating && (
                                          <Spinner className="text-black dark:text-white absolute" />
                                        )}
                                    </div>
                                  </div>
                                )}
                              </FormControl>
                              <div className="inline-flex">
                                <TestConnectionBadge
                                  validationResponse={
                                    destinationValidation[
                                      form.getValues(
                                        `destinations.${index}.connectionId`
                                      )
                                    ]?.response
                                  }
                                  connectionId={form.getValues(
                                    `destinations.${index}.connectionId`
                                  )}
                                  accountName={account?.name ?? ''}
                                />
                              </div>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                      </div>
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
                    <FormField
                      control={form.control}
                      name={`destinations.${index}.destinationOptions`}
                      render={({ field }) => (
                        <FormItem>
                          <FormControl>
                            <DestinationOptionsForm
                              connection={connections.find(
                                (c) =>
                                  c.id ==
                                  form.getValues().destinations[index]
                                    .connectionId
                              )}
                              value={field.value}
                              setValue={(newOpts) => {
                                field.onChange(newOpts);
                              }}
                              hideDynamoDbTableMappings={true}
                              destinationDetailsRecord={{}} // not used beacause we are hiding dynamodb table mappings
                              errors={destErrs[index]?.destinationOptions}
                            />
                          </FormControl>
                        </FormItem>
                      )}
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
