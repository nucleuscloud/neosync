'use client';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { useToast } from '@/components/ui/use-toast';
import { useGetAccountTemporalConfig } from '@/libs/hooks/useGetAccountTemporalConfig';
import {
  AccountTemporalConfig,
  GetAccountTemporalConfigResponse,
  SetAccountTemporalConfigRequest,
  SetAccountTemporalConfigResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { getErrorMessage } from '@/util/util';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import * as Yup from 'yup';

const FORM_SCHEMA = Yup.object({
  namespace: Yup.string().required(),
  syncJobName: Yup.string().required(),
  temporalUrl: Yup.string().required(),
});

type FormValues = Yup.InferType<typeof FORM_SCHEMA>;

export default function TemporalConfigForm(): ReactElement {
  const account = useAccount();
  const { data: tcData, mutate: mutateTcData } = useGetAccountTemporalConfig();
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
      const updateResp = await setTemporalConfig(values);
      mutateTcData(
        new GetAccountTemporalConfigResponse({
          config: updateResp.config,
        })
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

  return (
    <div>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-3">
          <FormField
            control={form.control}
            name="temporalUrl"
            render={({ field }) => (
              <FormItem>
                <FormControl>
                  <Input placeholder="Temporal URL" {...field} />
                </FormControl>
                <FormDescription>
                  The temporal url that will be used to connect to the temporal
                  instance.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="namespace"
            render={({ field }) => (
              <FormItem>
                <FormControl>
                  <Input placeholder="Temporal Namespace" {...field} />
                </FormControl>
                <FormDescription>
                  The name of the temporal namespace.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="syncJobName"
            render={({ field }) => (
              <FormItem>
                <FormControl>
                  <Input placeholder="Worker Sync Job Queue Name" {...field} />
                </FormControl>
                <FormDescription>
                  The name of the temporal job queue for the sync-job worker.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <Button type="submit">Submit</Button>
        </form>
      </Form>
    </div>
  );
}

async function setTemporalConfig(
  values: FormValues
): Promise<SetAccountTemporalConfigResponse> {
  const res = await fetch(`/api/users/accounts/temporal-config`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new SetAccountTemporalConfigRequest({
        config: new AccountTemporalConfig({
          namespace: values.namespace,
          syncJobQueueName: values.syncJobName,
          url: values.temporalUrl,
        }),
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return SetAccountTemporalConfigResponse.fromJson(await res.json());
}
