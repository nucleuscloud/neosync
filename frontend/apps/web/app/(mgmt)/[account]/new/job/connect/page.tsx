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
import { ReactElement, useEffect } from 'react';
import { UseFormReturn, useFieldArray, useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import DestinationOptionsForm from '../../../../../../components/jobs/Form/DestinationOptionsForm';
import JobsProgressSteps, { DATA_SYNC_STEPS } from '../JobsProgressSteps';
import { CONNECT_FORM_SCHEMA, ConnectFormValues } from '../schema';

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

  const sessionPrefix = searchParams?.sessionId ?? '';
  const sessionKey = `${sessionPrefix}-new-job-connect`;
  const [defaultValues] = useSessionStorage<ConnectFormValues>(sessionKey, {
    sourceId: '',
    sourceOptions: {},
    destinations: [{ connectionId: '', destinationOptions: {} }],
  });

  const form = useForm<ConnectFormValues>({
    mode: 'onChange',
    resolver: yupResolver<ConnectFormValues>(CONNECT_FORM_SCHEMA),
    values: defaultValues,
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
  const { isLoading: isConnectionsLoading, data: connectionsData } =
    useGetConnections(account?.id ?? '');

  const connections = connectionsData?.connections ?? [];

  async function onSubmit(_values: ConnectFormValues) {
    router.push(`/${account?.name}/new/job/schema?sessionId=${sessionPrefix}`);
  }

  const { postgres, mysql, s3 } = splitConnections(connections);

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
                            if (!value) {
                              return;
                            }
                            if (value === NEW_CONNECTION_VALUE) {
                              router.push(
                                `/${account?.name}/new/connection?returnTo=${encodeURIComponent(
                                  `/${account?.name}/new/job/connect?sessionId=${sessionPrefix}&from=new-connection`
                                )}`
                              );
                              return;
                            }
                            // set value
                            field.onChange(value);
                            form.setValue('sourceOptions', {
                              haltOnNewColumnAddition: false,
                            });
                            // clear form state for source+dst connection ids and re-evaluate the form state
                            handleConnectionValidation(form, connections);
                          }}
                          value={field.value}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Select a source ..." />
                          </SelectTrigger>
                          <SelectContent>
                            <ConnectionSelectContent
                              postgres={postgres}
                              mysql={mysql}
                              s3={[]}
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
              {fields.map((_, index) => {
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
                                      if (!value) {
                                        return;
                                      }
                                      if (value === NEW_CONNECTION_VALUE) {
                                        router.push(
                                          `/${account?.name}/new/connection?returnTo=${encodeURIComponent(
                                            `/${account?.name}/new/job/connect?sessionId=${sessionPrefix}&from=new-connection`
                                          )}`
                                        );
                                        return;
                                      }
                                      // set values
                                      form.setValue(
                                        `destinations.${index}.destinationOptions`,
                                        {
                                          truncateBeforeInsert: false,
                                          truncateCascade: false,
                                          initTableSchema: false,
                                        }
                                      );
                                      field.onChange(value);

                                      // clear form state for source+dst connection ids and re-evaluate the form state
                                      handleConnectionValidation(
                                        form,
                                        connections
                                      );
                                    }}
                                    value={field.value}
                                  >
                                    <SelectTrigger>
                                      <SelectValue placeholder="Select a destination ..." />
                                    </SelectTrigger>
                                    <SelectContent>
                                      <ConnectionSelectContent
                                        postgres={postgres}
                                        mysql={mysql}
                                        s3={s3}
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

function splitConnections(connections: Connection[]): {
  postgres: Connection[];
  mysql: Connection[];
  s3: Connection[];
} {
  const postgres: Connection[] = [];
  const mysql: Connection[] = [];
  const s3: Connection[] = [];

  connections.forEach((connection) => {
    if (connection.connectionConfig?.config.case === 'pgConfig') {
      postgres.push(connection);
    } else if (connection.connectionConfig?.config.case === 'mysqlConfig') {
      mysql.push(connection);
    } else if (connection.connectionConfig?.config.case === 'awsS3Config') {
      s3.push(connection);
    }
  });

  return {
    postgres,
    mysql,
    s3,
  };
}

interface ConnectionSelectContentProps {
  postgres: Connection[];
  mysql: Connection[];
  s3: Connection[];
}
function ConnectionSelectContent(
  props: ConnectionSelectContentProps
): ReactElement {
  const { postgres, mysql, s3 } = props;
  return (
    <>
      {postgres.length > 0 && (
        <SelectGroup>
          <SelectLabel>Postgres</SelectLabel>
          {postgres.map((connection) => (
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

      {mysql.length > 0 && (
        <SelectGroup>
          <SelectLabel>Mysql</SelectLabel>
          {mysql.map((connection) => (
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
      {s3.length > 0 && (
        <SelectGroup>
          <SelectLabel>AWS S3</SelectLabel>
          {s3.map((connection) => (
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
    </>
  );
}

function getErrorConnectionTypes(
  isSource: boolean,
  connId: string,
  connections: Connection[]
): string {
  const conn = connections.find((c) => c.id === connId);
  if (!conn) {
    return isSource ? '[Postgres, Mysql]' : '[Postgres, Mysql, AWS S3]';
  }
  if (conn.connectionConfig?.config.case === 'awsS3Config') {
    return '[Mysql, Postgres]';
  }
  if (conn.connectionConfig?.config.case === 'mysqlConfig') {
    return isSource ? '[Postgres]' : '[Mysql, AWS S3]';
  }

  if (conn.connectionConfig?.config.case === 'pgConfig') {
    return isSource ? '[Mysql]' : '[Postgres, AWS S3]';
  }
  return '';
}

function isValidConnectionPair(
  connId1: string,
  connId2: string,
  connections: Connection[]
): boolean {
  const conn1 = connections.find((c) => c.id === connId1);
  const conn2 = connections.find((c) => c.id === connId2);

  if (!conn1 || !conn2) {
    return true;
  }
  if (
    conn1.connectionConfig?.config.case === 'awsS3Config' ||
    conn2.connectionConfig?.config.case === 'awsS3Config'
  ) {
    return true;
  }

  if (
    conn1.connectionConfig?.config.case === conn2.connectionConfig?.config.case
  ) {
    return true;
  }

  return false;
}

function handleConnectionValidation(
  form: UseFormReturn<ConnectFormValues>,
  connections: Connection[]
): void {
  // clear source error before re-evaluation
  form.clearErrors('sourceId');

  const sourceId = form.getValues('sourceId');

  const destinationIds = form
    .getValues('destinations')
    .map((d) => d.connectionId);

  // initially reset the destinationIds state
  destinationIds.forEach((_, idx) =>
    form.clearErrors(`destinations.${idx}.connectionId`)
  );

  // check to make sure that source->destination are unique
  if (destinationIds.some((id) => id === sourceId)) {
    // for whatever reason this only propagates if the destination is what triggers the error
    // at one point it worked and then it mystically stopped working, but overall it seems to still get the point across
    // to the user
    // if you console log the form state, it does set the sourceId error, but then it will unset it for some reason
    form.setError(`sourceId`, {
      type: 'string',
      message: 'Source must be different from destination',
    });
    destinationIds.forEach((did, idx) => {
      if (did === sourceId) {
        form.setError(`destinations.${idx}.connectionId`, {
          type: 'string',
          message: 'Destination must be different from the source',
        });
      }
    });
    return;
  }

  // check to make sure all source->destination pairs of valid
  if (
    destinationIds.some(
      (did) => !isValidConnectionPair(sourceId, did, connections)
    )
  ) {
    // set the error on any of the destinations that are invalid
    destinationIds.forEach((did, idx) => {
      if (!isValidConnectionPair(sourceId, did, connections)) {
        form.setError(`destinations.${idx}.connectionId`, {
          type: 'string',
          message: `Destination connection type must one of ${getErrorConnectionTypes(
            false,
            sourceId,
            connections
          )}`,
        });
      }
    });
    return;
  }

  // check to make sure that all destination ids are unique
  if (destinationIds.length !== new Set(destinationIds).size) {
    const valueIdxMap = new Map<string, number[]>();
    destinationIds.forEach((dstId, idx) => {
      const idxs = valueIdxMap.get(dstId);
      if (idxs !== undefined) {
        idxs.push(idx);
      } else {
        valueIdxMap.set(dstId, [idx]);
      }
    });

    Array.from(valueIdxMap.values()).forEach((indices) => {
      if (indices.length > 1) {
        indices.forEach((idx) =>
          form.setError(`destinations.${idx}.connectionId`, {
            type: 'string',
            message: 'Destination connections must be unique from one another',
          })
        );
      }
    });
  }
}
