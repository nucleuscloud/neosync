'use client';
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
import { PlusIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import JobsProgressSteps, { DATA_GEN_STEPS } from '../../../JobsProgressSteps';
import {
  SINGLE_TABLE_CONNECT_FORM_SCHEMA,
  SingleTableConnectFormValues,
} from '../../../schema';

const NEW_CONNECTION_VALUE = 'new-connection';

export default function Page({ searchParams }: PageProps): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  useEffect(() => {
    if (!searchParams?.sessionId) {
      router.push(`/new/job`);
    }
  }, [searchParams?.sessionId]);

  const sessionPrefix = searchParams?.sessionId ?? '';
  const formKey = `${sessionPrefix}-new-job-single-table-connect`;
  const [defaultValues] = useSessionStorage<SingleTableConnectFormValues>(
    formKey,
    {
      connectionId: '',
      destinationOptions: {},
    }
  );

  const form = useForm({
    resolver: yupResolver<SingleTableConnectFormValues>(
      SINGLE_TABLE_CONNECT_FORM_SCHEMA
    ),
    defaultValues,
  });

  useFormPersist(formKey, {
    watch: form.watch,
    setValue: form.setValue,
    storage: window.sessionStorage,
  });
  const { isLoading: isConnectionsLoading, data: connectionsData } =
    useGetConnections(account?.id ?? '');

  const connections = connectionsData?.connections ?? [];

  async function onSubmit(_values: SingleTableConnectFormValues) {
    router.push(`/new/job/generate/single/schema?sessionId=${sessionPrefix}`);
  }

  const errors = form.formState.errors;

  const mysqlConns = connections.filter(
    (c) => c.connectionConfig?.config.case == 'mysqlConfig'
  );
  const postgresConns = connections.filter(
    (c) => c.connectionConfig?.config.case == 'pgConfig'
  );

  return (
    <div
      id="newjobflowcontainer"
      className="px-12 md:px-24 lg:px-32 flex flex-col gap-20"
    >
      <OverviewContainer
        Header={
          <PageHeader
            header="Connect"
            progressSteps={
              <JobsProgressSteps steps={DATA_GEN_STEPS} stepName={'connect'} />
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
                name="connectionId"
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
                                  `/new/job/generate/single/connect?sessionId=${sessionPrefix}`
                                )}`
                              );
                              return;
                            }
                            field.onChange(value);
                            form.setValue('destinationOptions', {
                              initTableSchema: false,
                              truncateBeforeInsert: false,
                              truncateCascade: false,
                            });
                          }}
                          value={field.value}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Select a connection ..." />
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

              <DestinationOptionsForm
                connection={connections.find(
                  (c) => c.id == form.getValues().connectionId
                )}
                maxColNum={2}
              />
            </div>
          </div>
          <Separator className="my-6" />

          <div className="flex flex-row gap-1 justify-between">
            <Button type="button" onClick={() => router.back()}>
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
