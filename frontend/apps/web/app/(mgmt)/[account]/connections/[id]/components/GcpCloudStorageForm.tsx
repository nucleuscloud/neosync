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
import {
  EditConnectionFormContext,
  GcpCloudStorageFormValues,
} from '@/yup-validations/connections';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { ConnectionService, UpdateConnectionResponse } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { buildConnectionConfigGcpCloudStorage } from '../../util';

interface Props {
  connectionId: string;
  defaultValues: GcpCloudStorageFormValues;
  onSaved(updatedResp: UpdateConnectionResponse): void;
  onSaveFailed(err: unknown): void;
}

export default function GcpCloudStorageForm(props: Props): ReactElement {
  const { connectionId, defaultValues, onSaved, onSaveFailed } = props;
  const { account } = useAccount();
  const { mutateAsync: isConnectionNameAvailableAsync } = useMutation(
    ConnectionService.method.isConnectionNameAvailable
  );
  const form = useForm<GcpCloudStorageFormValues, EditConnectionFormContext>({
    resolver: yupResolver(GcpCloudStorageFormValues),
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

  async function onSubmit(values: GcpCloudStorageFormValues): Promise<void> {
    if (!account) {
      return;
    }

    try {
      const connectionResp = await mutateAsync({
        id: connectionId,
        name: values.connectionName,
        connectionConfig: buildConnectionConfigGcpCloudStorage(values),
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

        <div className="flex flex-row gap-2 justify-end">
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
