'use client';
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
import { Separator } from '@/components/ui/separator';
import { Textarea } from '@/components/ui/textarea';
import {
  POSTGRES_FORM_SCHEMA,
  PostgresFormValues,
  SSL_MODES,
} from '@/yup-validations/connections';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckConnectionConfigResponse,
  ConnectionConfig,
  PostgresConnection,
  PostgresConnectionConfig,
  SSHAuthentication,
  SSHPassphrase,
  SSHPrivateKey,
  SSHTunnel,
  UpdateConnectionRequest,
  UpdateConnectionResponse,
} from '@neosync/sdk';
import {
  CheckCircledIcon,
  ExclamationTriangleIcon,
} from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';

interface Props {
  connectionId: string;
  defaultValues: PostgresFormValues;
  onSaved(updatedConnectionResp: UpdateConnectionResponse): void;
  onSaveFailed(err: unknown): void;
}

export default function PostgresForm(props: Props) {
  const { connectionId, defaultValues, onSaved, onSaveFailed } = props;
  const { account } = useAccount();
  const form = useForm<PostgresFormValues>({
    resolver: yupResolver(POSTGRES_FORM_SCHEMA),
    values: defaultValues,
    context: {
      originalConnectionName: defaultValues.connectionName,
      accountId: account?.id ?? '',
    }, // used when validating a new connection name
  });
  const [checkResp, setCheckResp] = useState<
    CheckConnectionConfigResponse | undefined
  >();

  const [isTesting, setIsTesting] = useState<boolean>(false);

  async function onSubmit(values: PostgresFormValues) {
    try {
      const connectionResp = await updatePostgresConnection(
        connectionId,
        values.connectionName,
        values.db,
        values.tunnel,
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
              <FormDescription>The database port.</FormDescription>
              <FormControl>
                <Input placeholder="5432" {...field} />
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
                <Input placeholder="postgres" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="db.user"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                Database Username
              </FormLabel>
              <FormDescription>The database username</FormDescription>
              <FormControl>
                <Input placeholder="postgres" {...field} />
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
                <Input placeholder="postgres" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="db.sslMode"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                SSL Mode
              </FormLabel>
              <FormDescription>
                Turn on SSL Mode to use TLS for client/server encryption.
              </FormDescription>
              <FormControl>
                <Select onValueChange={field.onChange} value={field.value}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {SSL_MODES.map((mode) => (
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

        <Separator />
        <div className="flex flex-col gap-2">
          <h2 className="text-xl font-semibold tracking-tight">
            Bastion Host Configuration
          </h2>
          <p>
            This section is optional and only necessary if your database is not
            publicly accessible to the internet.
          </p>
        </div>
        <FormField
          control={form.control}
          name="tunnel.host"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Host</FormLabel>
              <FormDescription>
                The hostname of the bastion server that will be used for SSH
                tunneling.
              </FormDescription>
              <FormControl>
                <Input placeholder="bastion.example.com" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="tunnel.port"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Port</FormLabel>
              <FormDescription>
                The port of the bastion host. Typically this is port 22.
              </FormDescription>
              <FormControl>
                <Input
                  type="number"
                  placeholder="22"
                  {...field}
                  onChange={(e) => {
                    field.onChange(e.target.valueAsNumber);
                  }}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="tunnel.user"
          render={({ field }) => (
            <FormItem>
              <FormLabel>User</FormLabel>
              <FormDescription>
                The name of the user that will be used to authenticate.
              </FormDescription>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="tunnel.privateKey"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Private Key</FormLabel>
              <FormDescription>
                The private key that will be used to authenticate against the
                SSH server. If using passphrase auth, provide that in the
                appropriate field below.
              </FormDescription>
              <FormControl>
                <Textarea {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="tunnel.passphrase"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Passphrase / Private Key Password</FormLabel>
              <FormDescription>
                The passphrase that will be used to authenticate with. If the
                SSH Key provided above is encrypted, provide the password for it
                here.
              </FormDescription>
              <FormControl>
                <Input type="password" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Separator />

        <TestConnectionResult resp={checkResp} />
        <div className="flex flex-row gap-3 justify-between">
          <Button
            variant="outline"
            disabled={!form.formState.isValid}
            onClick={async () => {
              setIsTesting(true);
              try {
                const resp = await checkPostgresConnection(
                  form.getValues('db'),
                  form.getValues('tunnel'),
                  account?.id ?? ''
                );
                setCheckResp(resp);
              } catch (err) {
                setCheckResp(
                  new CheckConnectionConfigResponse({
                    isConnected: false,
                    connectionError:
                      err instanceof Error ? err.message : 'unknown error',
                  })
                );
              } finally {
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

async function updatePostgresConnection(
  connectionId: string,
  connectionName: string,
  db: PostgresFormValues['db'],
  tunnel: PostgresFormValues['tunnel'],
  accountId: string
): Promise<UpdateConnectionResponse> {
  const pgconfig = new PostgresConnectionConfig({
    connectionConfig: {
      case: 'connection',
      value: new PostgresConnection({
        host: db.host,
        name: db.name,
        user: db.user,
        pass: db.pass,
        port: db.port,
        sslMode: db.sslMode,
      }),
    },
  });
  if (tunnel && tunnel.host) {
    pgconfig.tunnel = new SSHTunnel({
      host: tunnel.host,
      port: tunnel.port,
      user: tunnel.user,
      knownHostPublicKey: tunnel.knownHostPublicKey
        ? tunnel.knownHostPublicKey
        : undefined,
    });
    if (tunnel.privateKey) {
      pgconfig.tunnel.authentication = new SSHAuthentication({
        authConfig: {
          case: 'privateKey',
          value: new SSHPrivateKey({
            value: tunnel.privateKey,
            passphrase: tunnel.passphrase,
          }),
        },
      });
    } else if (tunnel.passphrase) {
      pgconfig.tunnel.authentication = new SSHAuthentication({
        authConfig: {
          case: 'passphrase',
          value: new SSHPassphrase({
            value: tunnel.passphrase,
          }),
        },
      });
    }
  }

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
              case: 'pgConfig',
              value: pgconfig,
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

async function checkPostgresConnection(
  db: PostgresFormValues['db'],
  tunnel: PostgresFormValues['tunnel'],
  accountId: string
): Promise<CheckConnectionConfigResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/connections/postgres/check`,
    {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify({ db, tunnel }),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CheckConnectionConfigResponse.fromJson(await res.json());
}
