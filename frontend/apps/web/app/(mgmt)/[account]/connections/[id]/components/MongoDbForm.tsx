'use client';
import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import RequiredLabel from '@/components/labels/RequiredLabel';
import PermissionsDialog from '@/components/permissions/PermissionsDialog';
import { useAccount } from '@/components/providers/account-provider';
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
import { MongoDbFormValues } from '@/yup-validations/connections';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckConnectionConfigResponse,
  UpdateConnectionResponse,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';
import { checkMongoConnection, updateMongoConnection } from '../../util';

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
  const [isValidating, setIsValidating] = useState<boolean>(false);
  const [validationResponse, setValidationResponse] = useState<
    CheckConnectionConfigResponse | undefined
  >();
  const [openPermissionDialog, setOpenPermissionDialog] =
    useState<boolean>(false);

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
      setOpenPermissionDialog(!!res.isConnected);
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
      const connectionResp = await updateMongoConnection(
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

        <PermissionsDialog
          checkResponse={
            validationResponse ?? new CheckConnectionConfigResponse({})
          }
          openPermissionDialog={openPermissionDialog}
          setOpenPermissionDialog={setOpenPermissionDialog}
          isValidating={isValidating}
          connectionName={form.getValues('connectionName')}
          connectionType="mongodb"
        />

        <div className="flex flex-row gap-2 justify-end">
          <Button
            type="button"
            variant="outline"
            onClick={() => onValidationClick()}
          >
            <ButtonText
              leftIcon={
                isValidating ? (
                  <Spinner className="text-black dark:text-white" />
                ) : (
                  <div />
                )
              }
              text="Test Connection"
            />
          </Button>

          <Button type="submit">
            <ButtonText
              leftIcon={form.formState.isSubmitting ? <Spinner /> : null}
              text="Update"
            />
          </Button>
        </div>
        {validationResponse && !validationResponse.isConnected && (
          <ErrorAlert
            title="Unable to connect"
            description={
              validationResponse.connectionError ?? 'no error returned'
            }
          />
        )}
      </form>
    </Form>
  );
}

interface ErrorAlertProps {
  title: string;
  description: string;
}
function ErrorAlert(props: ErrorAlertProps): ReactElement {
  const { title, description } = props;
  return (
    <Alert variant="destructive">
      <ExclamationTriangleIcon className="h-4 w-4" />
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription>{description}</AlertDescription>
    </Alert>
  );
}
