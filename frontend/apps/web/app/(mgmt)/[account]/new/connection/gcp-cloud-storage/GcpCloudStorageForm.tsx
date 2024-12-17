'use client';
import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import RequiredLabel from '@/components/labels/RequiredLabel';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { getErrorMessage } from '@/util/util';
import {
  CreateConnectionFormContext,
  GcpCloudStorageFormValues,
} from '@/yup-validations/connections';
import { create } from '@bufbuild/protobuf';
import { createConnectQueryKey, useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { GetConnectionResponseSchema } from '@neosync/sdk';
import {
  createConnection,
  getConnection,
  isConnectionNameAvailable,
} from '@neosync/sdk/connectquery';
import { useQueryClient } from '@tanstack/react-query';
import Error from 'next/error';
import { useRouter, useSearchParams } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { IoAlertCircleOutline } from 'react-icons/io5';
import { toast } from 'sonner';
import { buildConnectionConfigGcpCloudStorage } from '../../../connections/util';

export default function GcpCloudStorageForm(): ReactElement {
  const searchParams = useSearchParams();
  const { account } = useAccount();
  const sourceConnId = searchParams.get('sourceId');
  const [isLoading, setIsLoading] = useState<boolean>();
  const { data: systemAppConfig, isLoading: isSystemAppConfigLoading } =
    useGetSystemAppConfig();
  const { mutateAsync: isConnectionNameAvailableAsync } = useMutation(
    isConnectionNameAvailable
  );
  const form = useForm<GcpCloudStorageFormValues, CreateConnectionFormContext>({
    resolver: yupResolver(GcpCloudStorageFormValues),
    mode: 'onChange',
    defaultValues: {
      connectionName: '',
      gcp: {
        bucket: '',
        pathPrefix: '',
      },
    },
    context: {
      accountId: account?.id ?? '',
      isConnectionNameAvailable: isConnectionNameAvailableAsync,
    },
  });
  const router = useRouter();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const posthog = usePostHog();
  const { mutateAsync: createGcpCloudStorageConnection } =
    useMutation(createConnection);
  const { mutateAsync: getGcpCloudStorageConnection } =
    useMutation(getConnection);
  const queryclient = useQueryClient();

  async function onSubmit(values: GcpCloudStorageFormValues) {
    if (!account || isSubmitting) {
      return;
    }
    setIsSubmitting(true);
    try {
      const newConnection = await createGcpCloudStorageConnection({
        name: values.connectionName,
        accountId: account.id,
        connectionConfig: buildConnectionConfigGcpCloudStorage(values),
      });
      posthog.capture('New Connection Created', { type: 'gcp-cloud-storage' });
      toast.success('Successfully created connection!');
      const returnTo = searchParams.get('returnTo');
      if (returnTo) {
        router.push(returnTo);
      } else if (newConnection.connection?.id) {
        queryclient.setQueryData(
          createConnectQueryKey({
            schema: getConnection,
            input: { id: newConnection.connection.id },
            cardinality: undefined,
          }),
          create(GetConnectionResponseSchema, {
            connection: newConnection.connection,
          })
        );
        router.push(
          `/${account?.name}/connections/${newConnection.connection.id}`
        );
      } else {
        router.push(`/${account.name}/connections`);
      }
    } catch (err) {
      toast.error('Unable to create connection', {
        description: getErrorMessage(err),
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  useEffect(() => {
    const fetchData = async () => {
      if (!sourceConnId || !account?.id) {
        return;
      }
      setIsLoading(true);
      try {
        const connData = await getGcpCloudStorageConnection({
          id: sourceConnId,
        });
        if (
          connData.connection?.connectionConfig?.config.case !==
          'gcpCloudstorageConfig'
        ) {
          return;
        }

        const config = connData.connection?.connectionConfig?.config.value;

        form.reset({
          ...form.getValues(),
          connectionName: connData.connection?.name + '-copy',
          gcp: {
            ...config,
          },
        });
      } catch (error) {
        console.error('Failed to fetch connection data:', error);
        toast.error('Unable to retrieve connection data for clone!', {
          description: getErrorMessage(error),
        });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [account?.id]);

  if (isLoading || !account?.id || isSystemAppConfigLoading) {
    return <SkeletonForm />;
  }
  if (!systemAppConfig?.isGcpCloudStorageConnectionsEnabled) {
    return <Error statusCode={404} />;
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <Alert variant="warning">
          <div className="flex flex-row items-center gap-2">
            <IoAlertCircleOutline className="h-6 w-6" />
            <AlertTitle className="font-semibold">Heads up!</AlertTitle>
          </div>
          <AlertDescription className="pl-8">
            Right now GCP Cloud Storage connections can only be used as a
            destination
          </AlertDescription>
        </Alert>
        <FormField
          control={form.control}
          name="connectionName"
          render={({ field: { onChange, ...field } }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                Connection Name
              </FormLabel>
              <FormDescription>
                The unique name of the connection
              </FormDescription>
              <FormControl>
                <Input
                  placeholder="Connection Name"
                  {...field}
                  onChange={async ({ target: { value } }) => {
                    onChange(value);
                    await form.trigger('connectionName');
                  }}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="gcp.bucket"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                Bucket
              </FormLabel>
              <FormDescription>The bucket</FormDescription>
              <FormControl>
                <Input placeholder="Bucket" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="gcp.pathPrefix"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Path Prefix</FormLabel>
              <FormDescription>The path prefix of the bucket</FormDescription>
              <FormControl>
                <Input placeholder="/..." {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="flex flex-row gap-3 justify-end">
          <Button type="submit">
            <ButtonText
              leftIcon={form.formState.isSubmitting ? <Spinner /> : <div></div>}
              text="Submit"
            />
          </Button>
        </div>
      </form>
    </Form>
  );
}
