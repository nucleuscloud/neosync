'use client';
import FormError from '@/components/FormError';
import {
  getAccount,
  useAccount,
} from '@/components/providers/account-provider';
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
  CreateConnectionRequest,
  CreateConnectionResponse,
  IsConnectionNameAvailableResponse,
  PostgresConnection,
  PostgresConnectionConfig,
} from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { SSL_MODES } from '@/yup-validations/connections';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckCircledIcon,
  ExclamationTriangleIcon,
} from '@radix-ui/react-icons';
import { useRouter, useSearchParams } from 'next/navigation';
import { ReactElement, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import * as Yup from 'yup';

const FORM_SCHEMA = Yup.object({
  connectionName: Yup.string()
    .required('Connection Name is a required field')
    .test(
      'validConnectionName',
      'Checking Connection Name',
      async (value, context) => {
        if (!value || value.length < 3) {
          return context.createError({
            message: 'Connection Name must be at least 3 characters long',
          });
        }
        const account = getAccount();
        if (!account) {
          return false;
        }

        try {
          const res = await isConnectionNameAvailable(value, account.id);
          if (!res.isAvailable) {
            return context.createError({
              message: 'This Connection Name is already taken.',
            });
          }
          return true;
        } catch (error) {
          console.error('Error checking connection name availability:', error);
          return context.createError({
            message: 'Error validating name availability.',
          });
        }
      }
    ),
  db: Yup.object({
    host: Yup.string().required(),
    name: Yup.string().required(),
    user: Yup.string().required(),
    pass: Yup.string().required(),
    port: Yup.number().integer().positive().required(),
    sslMode: Yup.string().optional(),
  }).required(),
});

type FormValues = Yup.InferType<typeof FORM_SCHEMA>;

export default function PostgresForm() {
  const { account } = useAccount();
  const form = useForm<FormValues>({
    resolver: yupResolver(FORM_SCHEMA),
    defaultValues: {
      connectionName: '',
      db: {
        host: 'localhost',
        name: 'postgres',
        user: 'postgres',
        pass: 'postgres',
        port: 5432,
        sslMode: 'disable',
      },
    },
  });
  const router = useRouter();
  const searchParams = useSearchParams();
  const [checkResp, setCheckResp] = useState<
    CheckConnectionConfigResponse | undefined
  >();

  async function onSubmit(values: FormValues) {
    if (!account) {
      return;
    }
    try {
      const connection = await createPostgresConnection(
        values.db,
        values.connectionName,
        account.id
      );

      const returnTo = searchParams.get('returnTo');
      if (returnTo) {
        router.push(returnTo);
      } else if (connection.connection?.id) {
        router.push(`/connections/${connection.connection.id}`);
      } else {
        router.push(`/connections`);
      }
    } catch (err) {
      console.error(err);
    }
  }

  return (
    <div className="mx-64">
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <Controller
            control={form.control}
            name="connectionName"
            render={({ field: { onChange, ...field } }) => (
              <FormItem>
                <FormLabel>Connection Name</FormLabel>
                <FormDescription>
                  The unique name of the connection.
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
                <FormLabel>Host Name</FormLabel>
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
                <FormLabel>Port</FormLabel>
                <FormDescription>The port of the database</FormDescription>
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
                <FormLabel>Database Name</FormLabel>
                <FormDescription>The name of the database</FormDescription>
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
                <FormLabel>Database Username</FormLabel>
                <FormDescription>The username</FormDescription>
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
                <FormLabel>Database Password</FormLabel>
                <FormDescription>Password</FormDescription>
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
                <FormLabel>SSL Mode</FormLabel>
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
          <TestConnectionResult resp={checkResp} />
          <div className="flex flex-row gap-3 justify-between">
            <Button
              variant="outline"
              onClick={async () => {
                try {
                  const resp = await checkPostgresConnection(
                    form.getValues().db
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
                }
              }}
              type="button"
            >
              Test Connection
            </Button>
            <Button type="submit">Submit</Button>
          </div>
        </form>
      </Form>
    </div>
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
          description="Successfully connected to database!"
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
async function createPostgresConnection(
  db: FormValues['db'],
  name: string,
  accountId: string
): Promise<CreateConnectionResponse> {
  const res = await fetch(`/api/connections`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new CreateConnectionRequest({
        accountId,
        name: name,
        connectionConfig: new ConnectionConfig({
          config: {
            case: 'pgConfig',
            value: new PostgresConnectionConfig({
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

async function checkPostgresConnection(
  db: FormValues['db']
): Promise<CheckConnectionConfigResponse> {
  const res = await fetch(`/api/connections/postgres/check`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(db),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CheckConnectionConfigResponse.fromJson(await res.json());
}

export async function isConnectionNameAvailable(
  name: string,
  accountId: string
): Promise<IsConnectionNameAvailableResponse> {
  const res = await fetch(
    `/api/connections/is-connection-name-available?connectionName=${name}&accountId=${accountId}`,
    {
      method: 'GET',
      headers: {
        'content-type': 'application/json',
      },
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return IsConnectionNameAvailableResponse.fromJson(await res.json());
}
