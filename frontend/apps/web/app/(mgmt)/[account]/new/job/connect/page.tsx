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
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { yupResolver } from '@hookform/resolvers/yup';
import { Connection } from '@neosync/sdk';
import { Cross2Icon, PlusIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';
import { useFieldArray, useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import DestinationOptionsForm from '../../../../../../components/jobs/Form/DestinationOptionsForm';
import JobsProgressSteps, { DATA_SYNC_STEPS } from '../JobsProgressSteps';
import { CONNECT_FORM_SCHEMA, ConnectFormValues } from '../schema';

const NEW_CONNECTION_VALUE = 'new-connection';

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/${account?.name}/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = searchParams?.sessionId ?? '';
  const [defaultValues] = useSessionStorage<ConnectFormValues>(
    `${sessionPrefix}-new-job-connect`,
    {
      sourceId: '',
      sourceOptions: {},
      destinations: [{ connectionId: '', destinationOptions: {} }],
    }
  );

  const form = useForm({
    resolver: yupResolver<ConnectFormValues>(CONNECT_FORM_SCHEMA),
    defaultValues,
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: 'destinations',
  });

  useFormPersist(`${sessionPrefix}-new-job-connect`, {
    watch: form.watch,
    setValue: form.setValue,
    storage: window.sessionStorage,
  });
  const { isLoading: isConnectionsLoading, data: connectionsData } =
    useGetConnections(account?.id ?? '');

  const connections = connectionsData?.connections ?? [];

  async function onSubmit(_values: ConnectFormValues) {
    router.push(`/${account?.name}/new/job/schema?sessionId=${sessionPrefix}`);
  }

  const [sourceConn, setSourceConn] = useState<string>(
    form.getValues().sourceId
  );

  const [destConn, setDestConn] = useState<string>(form.getValues().sourceId);

  const errors = form.formState.errors;

  const mysqlConns = connections.filter(
    (c) => c.connectionConfig?.config.case == 'mysqlConfig'
  );
  const postgresConns = connections.filter(
    (c) => c.connectionConfig?.config.case == 'pgConfig'
  );
  const awsS3Conns = connections.filter(
    (c) => c.connectionConfig?.config.case == 'awsS3Config'
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
              <JobsProgressSteps steps={DATA_SYNC_STEPS} stepName={'connect'} />
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
                          onValueChange={(value: string) => {
                            if (value === NEW_CONNECTION_VALUE) {
                              router.push(
                                `/${account?.name}/new/connection?returnTo=${encodeURIComponent(
                                  `/${account?.name}/new/job/connect?sessionId=${sessionPrefix}`
                                )}`
                              );
                              return;
                            }
                            if (value == destConn) {
                              form.setError(`sourceId`, {
                                type: 'string',
                                message:
                                  'Source must be different from destination',
                              });
                            }
                            if (
                              !isValidConnectionPair(
                                value,
                                destConn,
                                connections
                              )
                            ) {
                              const isSource = true;
                              form.setError(`sourceId`, {
                                type: 'string',
                                message: `Source connection type must one of ${getErrorConnectionTypes(
                                  isSource,
                                  value,
                                  connections
                                )}`,
                              });
                            }

                            if (
                              isValidConnectionPair(
                                value,
                                destConn,
                                connections
                              ) &&
                              value != destConn
                            ) {
                              form.clearErrors();
                            }
                            setSourceConn(value);
                            if (form.formState.errors) {
                              form.clearErrors;
                            }
                            field.onChange(value);
                            form.setValue('sourceOptions', {
                              haltOnNewColumnAddition: false,
                            });
                          }}
                          value={field.value}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Select a source ..." />
                          </SelectTrigger>
                          <SelectContent>
                            {postgresConns.length > 0 && (
                              <SelectGroup>
                                <SelectLabel>Postgres</SelectLabel>
                                {postgresConns.map((connection) => (
                                  <SelectItem
                                    className="cursor-pointer"
                                    key={connection.id}
                                    value={connection.id}
                                  >
                                    {connection.name}
                                  </SelectItem>
                                ))}
                              </SelectGroup>
                            )}

                            {mysqlConns.length > 0 && (
                              <SelectGroup>
                                <SelectLabel>Mysql</SelectLabel>
                                {mysqlConns.map((connection) => (
                                  <SelectItem
                                    className="cursor-pointer"
                                    key={connection.id}
                                    value={connection.id}
                                  >
                                    {connection.name}
                                  </SelectItem>
                                ))}
                              </SelectGroup>
                            )}
                            <SelectItem
                              className="cursor-pointer"
                              key="new-src-connection"
                              value={NEW_CONNECTION_VALUE}
                            >
                              <div className="flex flex-row gap-1 items-center">
                                <PlusIcon />
                                <p>New Connection</p>
                              </div>
                            </SelectItem>
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
                  (c) => c.id == form.getValues().sourceId
                )}
                maxColNum={2}
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
              {fields.map(({}, index) => {
                return (
                  <div className="space-y-4 col-span-2" key={index}>
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
                                    onValueChange={(value: string) => {
                                      if (value === NEW_CONNECTION_VALUE) {
                                        router.push(
                                          `/new/connection?returnTo=${encodeURIComponent(
                                            `/new/job/connect?sessionId=${sessionPrefix}`
                                          )}`
                                        );
                                        return;
                                      }
                                      setDestConn(value);
                                      if (value == sourceConn) {
                                        form.setError(
                                          `destinations.${index}.connectionId`,
                                          {
                                            type: 'string',
                                            message:
                                              'Destination must be different from source',
                                          }
                                        );
                                      }
                                      if (
                                        !isValidConnectionPair(
                                          value,
                                          sourceConn,
                                          connections
                                        )
                                      ) {
                                        const isSource = false;
                                        form.setError(
                                          `destinations.${index}.connectionId`,
                                          {
                                            type: 'string',
                                            message: `Destination connection type must one of ${getErrorConnectionTypes(
                                              isSource,
                                              sourceConn,
                                              connections
                                            )}`,
                                          }
                                        );
                                      }
                                      if (
                                        isValidConnectionPair(
                                          value,
                                          sourceConn,
                                          connections
                                        ) &&
                                        value != sourceConn
                                      ) {
                                        form.clearErrors();
                                      }
                                      form.setValue(
                                        `destinations.${index}.destinationOptions`,
                                        {
                                          truncateBeforeInsert: false,
                                          truncateCascade: false,
                                          initTableSchema: false,
                                        }
                                      );
                                      field.onChange(value);
                                    }}
                                    value={field.value}
                                  >
                                    <SelectTrigger>
                                      <SelectValue placeholder="Select a destination ..." />
                                    </SelectTrigger>
                                    <SelectContent>
                                      {postgresConns.length > 0 && (
                                        <SelectGroup>
                                          <SelectLabel>Postgres</SelectLabel>
                                          {postgresConns.map((connection) => (
                                            <SelectItem
                                              className="cursor-pointer"
                                              key={connection.id}
                                              value={connection.id}
                                            >
                                              {connection.name}
                                            </SelectItem>
                                          ))}
                                        </SelectGroup>
                                      )}

                                      {mysqlConns.length > 0 && (
                                        <SelectGroup>
                                          <SelectLabel>Mysql</SelectLabel>
                                          {mysqlConns.map((connection) => (
                                            <SelectItem
                                              className="cursor-pointer"
                                              key={connection.id}
                                              value={connection.id}
                                            >
                                              {connection.name}
                                            </SelectItem>
                                          ))}
                                        </SelectGroup>
                                      )}
                                      {awsS3Conns.length > 0 && (
                                        <SelectGroup>
                                          <SelectLabel>AWS S3</SelectLabel>
                                          {awsS3Conns.map((connection) => (
                                            <SelectItem
                                              className="cursor-pointer"
                                              key={connection.id}
                                              value={connection.id}
                                            >
                                              {connection.name}
                                            </SelectItem>
                                          ))}
                                        </SelectGroup>
                                      )}
                                      <SelectItem
                                        className="cursor-pointer"
                                        key="new-dst-connection"
                                        value={NEW_CONNECTION_VALUE}
                                      >
                                        <div className="flex flex-row gap-1 items-center">
                                          <PlusIcon />
                                          <p>New Connection</p>
                                        </div>
                                      </SelectItem>
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
                          disabled={fields.length == 1}
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
                      index={index}
                      connection={connections.find(
                        (c) =>
                          c.id ==
                          form.getValues().destinations[index].connectionId
                      )}
                      maxColNum={2}
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
            <Button type="button" onClick={() => router.back()}>
              Back
            </Button>
            <Button
              type="submit"
              disabled={
                (errors?.destinations?.length ?? 0) > 0 ||
                (errors.sourceId?.message?.length ?? 0) > 0
              }
            >
              Next
            </Button>
          </div>
        </form>
      </Form>
    </div>
  );
}

function getErrorConnectionTypes(
  isSource: boolean,
  connId: string,
  connections: Connection[]
): string {
  const conn = connections.find((c) => c.id == connId);
  if (!conn) {
    return isSource ? '[Postgres, Mysql]' : '[Postgres, Mysql, AWS S3]';
  }
  if (conn.connectionConfig?.config.case == 'awsS3Config') {
    return '[Mysql, Postgres]';
  }
  if (conn.connectionConfig?.config.case == 'mysqlConfig') {
    return isSource ? '[Postgres]' : '[Mysql, AWS S3]';
  }

  if (conn.connectionConfig?.config.case == 'pgConfig') {
    return isSource ? '[Mysql]' : '[Postgres, AWS S3]';
  }
  return '';
}

function isValidConnectionPair(
  connId1: string,
  connId2: string,
  connections: Connection[]
): boolean {
  const conn1 = connections.find((c) => c.id == connId1);
  const conn2 = connections.find((c) => c.id == connId2);

  if (!conn1 || !conn2) {
    return true;
  }
  if (
    conn1.connectionConfig?.config.case == 'awsS3Config' ||
    conn2.connectionConfig?.config.case == 'awsS3Config'
  ) {
    return true;
  }

  if (
    conn1.connectionConfig?.config.case == conn2.connectionConfig?.config.case
  ) {
    return true;
  }

  return false;
}
