'use client';
import ButtonText from '@/components/ButtonText';
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
import { MongoDbFormValues } from '@/yup-validations/connections';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckConnectionConfigResponse,
  ConnectionConfig,
  MongoConnectionConfig,
  UpdateConnectionRequest,
  UpdateConnectionResponse,
} from '@neosync/sdk';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';
import { checkMongoConnection } from '../../../new/connection/mongodb/MongoDBForm';

interface Props {
  connectionId: string;
  defaultValues: MongoDbFormValues;
  onSaved(updatedResp: UpdateConnectionResponse): void;
  onSaveFailed(err: unknown): void;
}

export default function MongoDbForm(props: Props): ReactElement {
  const { connectionId, defaultValues, onSaved, onSaveFailed } = props;
  const { account } = useAccount();

  const form = useForm<MongoDbFormValues>({
    resolver: yupResolver(MongoDbFormValues),
    mode: 'onChange',
    values: defaultValues,
    context: {
      originalConnectionName: defaultValues.connectionName,
      accountId: account?.id ?? '',
    },
  });
  const [isValidating, setIsValidating] = useState<boolean>();
  // todo
  const [validationResponse, setValidationResponse] = useState<
    CheckConnectionConfigResponse | undefined
  >();

  async function onValidationClick(): Promise<void> {
    if (isValidating) {
      return;
    }
    setIsValidating(true);
    try {
      const res = await checkMongoConnection(
        form.getValues(),
        account?.id ?? ''
      );
      setValidationResponse(res);
    } catch (err) {
      setValidationResponse(
        new CheckConnectionConfigResponse({
          isConnected: false,
          connectionError: err instanceof Error ? err.message : 'unknown error',
        })
      );
    } finally {
      setIsValidating(false);
    }
  }

  async function onSubmit(values: MongoDbFormValues): Promise<void> {
    if (!account) {
      return;
    }

    try {
      const connectionResp = await updateConnection(
        values,
        connectionId,
        account.id
      );
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
          name="url"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                Connection Url
              </FormLabel>
              <FormDescription>The url of the MongoDB server</FormDescription>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="flex flex-row gap-2 justify-end">
          <Button
            type="button"
            variant="outline"
            onClick={() => onValidationClick()}
          >
            <ButtonText leftIcon={<div></div>} text="Test Connection" />
          </Button>

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

async function updateConnection(
  values: MongoDbFormValues,
  connectionId: string,
  accountId: string
): Promise<UpdateConnectionResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/connections/${connectionId}`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(
        new UpdateConnectionRequest({
          id: connectionId,
          name: values.connectionName,
          connectionConfig: new ConnectionConfig({
            config: {
              case: 'mongoConfig',
              value: new MongoConnectionConfig({
                connectionConfig: {
                  case: 'url',
                  value: values.url,
                },
                clientTls: undefined,
                tunnel: undefined,
              }),
            },
          }),
        })
      ),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateConnectionResponse.fromJson(await res.json());
}
