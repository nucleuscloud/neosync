'use client';
import { checkMysqlConnection } from '@/app/[account]/new/connection/mysql/MysqlForm';
import ButtonText from '@/components/ButtonText';
import FormError from '@/components/FormError';
import Spinner from '@/components/Spinner';
import RequiredLabel from '@/components/labels/RequiredLabel';
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  CheckConnectionConfigResponse,
  ConnectionConfig,
  MysqlConnection,
  MysqlConnectionConfig,
  UpdateConnectionRequest,
  UpdateConnectionResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import {
  MYSQL_CONNECTION_PROTOCOLS,
  MYSQL_FORM_SCHEMA,
  MysqlFormValues,
} from '@/yup-validations/connections';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckCircledIcon,
  ExclamationTriangleIcon,
} from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';

interface Props {
  connectionId: string;
  defaultValues: MysqlFormValues;
  onSaved(updatedConnectionResp: UpdateConnectionResponse): void;
  onSaveFailed(err: unknown): void;
}

export default function MysqlForm(props: Props) {
  const { connectionId, defaultValues, onSaved, onSaveFailed } = props;
  const form = useForm<MysqlFormValues>({
    resolver: yupResolver(MYSQL_FORM_SCHEMA),
    values: defaultValues,
    context: { originalConnectionName: defaultValues.connectionName },
  });
  const [checkResp, setCheckResp] = useState<
    CheckConnectionConfigResponse | undefined
  >();
  const { account } = useAccount();

  const [isTesting, setIsTesting] = useState<boolean>(false);

  async function onSubmit(values: MysqlFormValues) {
    try {
      const connectionResp = await updateMysqlConnection(
        connectionId,
        values.connectionName,
        values.db,
        account?.id ?? ''
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
        <Controller
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
              <FormError
                errorMessage={
                  form.formState.errors.connectionName?.message ?? ''
                }
              />
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="db.host"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                Host Name
              </FormLabel>
              <FormDescription>The host name</FormDescription>
              <FormControl>
                <Input placeholder="Host" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="db.port"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                Database Port
              </FormLabel>
              <FormDescription>The database port</FormDescription>
              <FormControl>
                <Input placeholder="3306" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="db.name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                Database Name
              </FormLabel>
              <FormDescription>The database name</FormDescription>
              <FormControl>
                <Input placeholder="mysql" {...field} />
              </FormControl>
              <FormDescription>The name of the database</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="db.user"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Database Username</FormLabel>
              <FormDescription>The username of the database</FormDescription>
              <FormControl>
                <Input placeholder="mysql" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="db.pass"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                Database Password
              </FormLabel>
              <FormDescription>The database password</FormDescription>
              <FormControl>
                <Input placeholder="mysql" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="db.protocol"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                Connection Protocol
              </FormLabel>
              <FormDescription>
                The protocol that you want to use to connect to your database
              </FormDescription>
              <FormControl>
                <Select onValueChange={field.onChange} value={field.value}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {MYSQL_CONNECTION_PROTOCOLS.map((mode) => (
                      <SelectItem
                        className="cursor-pointer"
                        key={mode}
                        value={mode}
                      >
                        {mode}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <TestConnectionResult resp={checkResp} />
        <div className="flex flex-row gap-3 justify-between">
          <Button
            variant="outline"
            disabled={!form.formState.isValid}
            onClick={async () => {
              setIsTesting(true);
              try {
                const resp = await checkMysqlConnection(form.getValues('db'));
                setCheckResp(resp);
                setIsTesting(false);
              } catch (err) {
                setCheckResp(
                  new CheckConnectionConfigResponse({
                    isConnected: false,
                    connectionError:
                      err instanceof Error ? err.message : 'unknown error',
                  })
                );
                setIsTesting(false);
              }
            }}
            type="button"
          >
            <ButtonText
              leftIcon={
                isTesting ? <Spinner className="text-black" /> : <div></div>
              }
              text="Test Connection"
            />
          </Button>

          <Button type="submit" disabled={!form.formState.isValid}>
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

interface TestConnectionResultProps {
  resp: CheckConnectionConfigResponse | undefined;
}

function TestConnectionResult(props: TestConnectionResultProps): ReactElement {
  const { resp } = props;
  if (resp) {
    if (resp.isConnected) {
      return (
        <SuccessAlert
          title="Success!"
          description="Successfully connected to the database!"
        />
      );
    } else {
      return (
        <ErrorAlert
          title="Unable to connect"
          description={resp.connectionError ?? 'no error returned'}
        />
      );
    }
  }
  return <div />;
}

interface SuccessAlertProps {
  title: string;
  description: string;
}

function SuccessAlert(props: SuccessAlertProps): ReactElement {
  const { title, description } = props;
  return (
    <Alert variant="success">
      <CheckCircledIcon className="h-4 w-4" />
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription>{description}</AlertDescription>
    </Alert>
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

async function updateMysqlConnection(
  connectionId: string,
  connectionName: string,
  db: MysqlFormValues['db'],
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
          name: connectionName,
          connectionConfig: new ConnectionConfig({
            config: {
              case: 'mysqlConfig',
              value: new MysqlConnectionConfig({
                connectionConfig: {
                  case: 'connection',
                  value: new MysqlConnection({
                    host: db.host,
                    name: db.name,
                    user: db.user,
                    pass: db.pass,
                    port: db.port,
                    protocol: db.protocol,
                  }),
                },
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
