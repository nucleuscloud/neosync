'use client';
import RequiredLabel from '@/components/labels/RequiredLabel';
import { PasswordInput } from '@/components/PasswordComponent';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
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
import { useToast } from '@/components/ui/use-toast';
import {
  buildGetConnectionRouteKey,
  getConnection,
} from '@/libs/hooks/useGetConnection';
import { getErrorMessage } from '@/util/util';
import { OpenAiFormValues } from '@/yup-validations/connections';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  ConnectionConfig,
  CreateConnectionRequest,
  CreateConnectionResponse,
  GetConnectionResponse,
  OpenAiConnectionConfig,
} from '@neosync/sdk';
import { ExternalLinkIcon } from '@radix-ui/react-icons';
import NextLink from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { mutate } from 'swr';

interface Props {}

export default function OpenAiForm(props: Props): ReactElement {
  const {} = props;
  const { account } = useAccount();
  const { toast } = useToast();
  const searchParams = useSearchParams();
  const sourceConnId = searchParams.get('sourceId');
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const router = useRouter();

  const form = useForm<OpenAiFormValues>({
    resolver: yupResolver(OpenAiFormValues),
    mode: 'onChange',
    defaultValues: {
      connectionName: '',
      sdk: {
        url: 'https://api.openai.com/v1',
        apiKey: '',
      },
    },
    context: { accountId: account?.id ?? '' },
  });

  async function onSubmit(values: OpenAiFormValues): Promise<void> {
    if (!account) {
      return;
    }

    try {
      const connectionResp = await createOpenAiConnection(account.id, values);
      toast({
        title: 'Successfully created OpenAI Connection!',
        variant: 'success',
      });

      mutate(
        buildGetConnectionRouteKey(
          account.id,
          connectionResp.connection?.id ?? ''
        ),
        new GetConnectionResponse({
          connection: connectionResp.connection,
        })
      );
      const returnTo = searchParams.get('returnTo');
      if (returnTo) {
        router.push(returnTo);
      } else if (connectionResp.connection?.id) {
        router.push(
          `/${account.name}/connections/${connectionResp.connection.id}`
        );
      }
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to create OpenAI Connection',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  /* we call the underlying useGetConnection API directly since we can't call
the hook in the useEffect conditionally. This is used to retrieve the values for the clone connection so that we can update the form.
*/
  useEffect(() => {
    const fetchData = async () => {
      if (!sourceConnId || !account?.id) {
        return;
      }
      setIsLoading(true);
      try {
        const connData = await getConnection(account.id, sourceConnId);

        if (
          connData &&
          connData.connection?.connectionConfig?.config.case === 'openaiConfig'
        ) {
          const config = connData.connection?.connectionConfig?.config.value;

          form.reset({
            ...form.getValues(),
            connectionName: connData.connection?.name + '-copy',
            sdk: {
              url: config.apiUrl,
              apiKey: config.apiKey,
            },
          });
        }
      } catch (error) {
        console.error('Failed to fetch connection data:', error);
        setIsLoading(false);
        toast({
          title: 'Unable to clone connection!',
          variant: 'destructive',
        });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [account?.id]);

  if (isLoading || !account?.id) {
    return <SkeletonForm />;
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <FormField
          control={form.control}
          name="connectionName"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                Connection Name
              </FormLabel>
              <FormDescription>
                The unique name of the connection
              </FormDescription>
              <FormControl>
                <Input placeholder="Connection Name" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="sdk.url"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                OpenAI API Url
              </FormLabel>
              <FormDescription>
                The url of the OpenAI API (or equivalent) server
              </FormDescription>
              <FormControl>
                <Input placeholder="https://api.openai.com/v1" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="sdk.apiKey"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                OpenAI API Key
              </FormLabel>
              <FormDescription>
                The api key to the API server. If you do not know how to create
                an OpenAI Key, navigate to their{' '}
                <NextLink
                  href="https://platform.openai.com/docs/quickstart/account-setup"
                  target="_blank"
                >
                  <div className="inline-flex gap-1 flex-row">
                    <p>Account Setup Docs</p>
                    <ExternalLinkIcon className="text-gray-800 w-4 h-4" />
                  </div>
                </NextLink>{' '}
                to learn how to do so.{' '}
              </FormDescription>
              <FormControl>
                <PasswordInput placeholder="Your api key here" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="flex flex-row justify-end">
          <Button type="submit">Submit</Button>
        </div>
      </form>
    </Form>
  );
}

async function createOpenAiConnection(
  accountId: string,
  values: OpenAiFormValues
): Promise<CreateConnectionResponse> {
  const res = await fetch(`/api/accounts/${accountId}/connections`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new CreateConnectionRequest({
        accountId,
        name: values.connectionName,
        connectionConfig: new ConnectionConfig({
          config: {
            case: 'openaiConfig',
            value: new OpenAiConnectionConfig({
              apiUrl: values.sdk.url,
              apiKey: values.sdk.apiKey,
            }),
          },
        }),
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CreateConnectionResponse.fromJson(await res.json());
}
