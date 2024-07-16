'use client';
import { useAccount } from '@/components/providers/account-provider';
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
import { Skeleton } from '@/components/ui/skeleton';
import { useToast } from '@/components/ui/use-toast';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { getErrorMessage } from '@/util/util';
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  AccountTemporalConfig,
  GetAccountTemporalConfigResponse,
} from '@neosync/sdk';
import {
  getAccountTemporalConfig,
  setAccountTemporalConfig,
} from '@neosync/sdk/connectquery';
import { useQueryClient } from '@tanstack/react-query';
import Error from 'next/error';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';

const FORM_SCHEMA = Yup.object({
  namespace: Yup.string().required(),
  syncJobName: Yup.string().required(),
  temporalUrl: Yup.string().required(),
});

type FormValues = Yup.InferType<typeof FORM_SCHEMA>;

export default function Temporal(): ReactElement {
  const { account } = useAccount();
  const { data: systemAppConfigData, isLoading: isSystemAppConfigDataLoading } =
    useGetSystemAppConfig();
  const { data: tcData, isLoading: isTemporalConfigLoading } = useQuery(
    getAccountTemporalConfig,
    { accountId: account?.id ?? '' },
    { enabled: !!account?.id }
  );
  const { mutateAsync } = useMutation(setAccountTemporalConfig);
  const queryclient = useQueryClient();

  const form = useForm<FormValues>({
    resolver: yupResolver(FORM_SCHEMA),
    defaultValues: {
      namespace: 'default',
      syncJobName: 'sync-job',
      temporalUrl: 'localhost:7233',
    },
    values: {
      namespace: tcData?.config?.namespace ?? 'default',
      syncJobName: tcData?.config?.syncJobQueueName ?? 'sync-job',
      temporalUrl: tcData?.config?.url ?? 'localhost:7233',
    },
  });
  const { toast } = useToast();
  async function onSubmit(values: FormValues): Promise<void> {
    if (!account) {
      return;
    }
    try {
      const updatedResp = await mutateAsync({
        accountId: account.id,
        config: new AccountTemporalConfig({
          namespace: values.namespace,
          syncJobQueueName: values.syncJobName,
          url: values.temporalUrl,
        }),
      });
      const key = createConnectQueryKey(getAccountTemporalConfig, {
        accountId: account.id,
      });
      queryclient.setQueryData(
        key,
        new GetAccountTemporalConfigResponse({ config: updatedResp.config })
      );
      toast({
        title: 'Successfully updated temporal config',
        variant: 'success',
      });
    } catch (err) {
      toast({
        title: 'Unable to submit temporal config',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }
  if (isSystemAppConfigDataLoading) {
    return <Skeleton />;
  }

  if (systemAppConfigData?.isNeosyncCloud) {
    return <Error statusCode={404} />;
  }

  if (isTemporalConfigLoading) {
    return <Skeleton />;
  }

  return (
    <div>
      <h1 className="text-xl font-bold tracking-tight">Temporal</h1>
      <Form {...form}>
        <form
          onSubmit={form.handleSubmit(onSubmit)}
          className="space-y-6 mt-10"
        >
          <div className="space-y-10 ">
            <FormField
              control={form.control}
              name="temporalUrl"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Temporal Instance URL</FormLabel>
                  <FormDescription>
                    The temporal url that will be used to connect to the
                    temporal instance.
                  </FormDescription>
                  <FormControl>
                    <Input placeholder="Temporal URL" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="namespace"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Temporal Namespace</FormLabel>
                  <FormDescription>
                    The name of the temporal namespace.
                  </FormDescription>
                  <FormControl>
                    <Input placeholder="Temporal Namespace" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="syncJobName"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Temporal Job Queue Name</FormLabel>
                  <FormDescription>
                    The name of the temporal job queue for the sync-job worker.
                  </FormDescription>
                  <FormControl>
                    <Input
                      placeholder="Worker Sync Job Queue Name"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
          <div className="flex justify-end">
            <Button type="submit">Submit</Button>
          </div>
        </form>
      </Form>
    </div>
  );
}
