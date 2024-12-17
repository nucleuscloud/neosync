'use client';
import ButtonText from '@/components/ButtonText';
import { PasswordInput } from '@/components/PasswordComponent';
import Spinner from '@/components/Spinner';
import RequiredLabel from '@/components/labels/RequiredLabel';
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
import {
  EditConnectionFormContext,
  OpenAiFormValues,
} from '@/yup-validations/connections';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { ConnectionService, UpdateConnectionResponse } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { buildConnectionConfigOpenAi } from '../../util';

interface Props {
  connectionId: string;
  defaultValues: OpenAiFormValues;
  onSaved(updatedResp: UpdateConnectionResponse): void;
  onSaveFailed(err: unknown): void;
}

export default function OpenAiForm(props: Props): ReactElement {
  const { connectionId, defaultValues, onSaved, onSaveFailed } = props;
  const { account } = useAccount();
  const { mutateAsync: isConnectionNameAvailableAsync } = useMutation(
    ConnectionService.method.isConnectionNameAvailable
  );
  const form = useForm<OpenAiFormValues, EditConnectionFormContext>({
    resolver: yupResolver(OpenAiFormValues),
    mode: 'onChange',
    values: defaultValues,
    context: {
      originalConnectionName: defaultValues.connectionName,
      accountId: account?.id ?? '',
      isConnectionNameAvailable: isConnectionNameAvailableAsync,
    },
  });
  const { mutateAsync } = useMutation(
    ConnectionService.method.updateConnection
  );

  async function onSubmit(values: OpenAiFormValues): Promise<void> {
    if (!account) {
      return;
    }

    try {
      const connectionResp = await mutateAsync({
        id: connectionId,
        name: values.connectionName,
        connectionConfig: buildConnectionConfigOpenAi(values),
      });
      onSaved(connectionResp);
    } catch (err) {
      console.error(err);
      onSaveFailed(err);
    }
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
              <FormDescription>The api key to the API server</FormDescription>
              <FormControl>
                <PasswordInput placeholder="Your api key here" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="flex flex-row justify-end">
          <Button type="submit">
            <ButtonText
              leftIcon={form.formState.isSubmitting ? <Spinner /> : null}
              text="Update"
            />
          </Button>
        </div>
      </form>
    </Form>
  );
}
