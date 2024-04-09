'use client';
import { checkMysqlConnection } from '@/app/(mgmt)/[account]/new/connection/mysql/MysqlForm';
import ButtonText from '@/components/ButtonText';
import FormError from '@/components/FormError';
import Spinner from '@/components/Spinner';
import RequiredLabel from '@/components/labels/RequiredLabel';
import { useAccount } from '@/components/providers/account-provider';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
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
import { Textarea } from '@/components/ui/textarea';
import {
  MYSQL_CONNECTION_PROTOCOLS,
  MYSQL_FORM_SCHEMA,
  MysqlFormValues,
} from '@/yup-validations/connections';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckConnectionConfigResponse,
  ConnectionConfig,
  ConnectionRolePrivilege,
  MysqlConnection,
  MysqlConnectionConfig,
  SSHAuthentication,
  SSHPassphrase,
  SSHPrivateKey,
  SSHTunnel,
  SqlConnectionOptions,
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
  defaultValues: MysqlFormValues;
  onSaved(updatedConnectionResp: UpdateConnectionResponse): void;
  onSaveFailed(err: unknown): void;
}

export default function MysqlForm(props: Props) {
  const { connectionId, defaultValues, onSaved, onSaveFailed } = props;
  const { account } = useAccount();
  const form = useForm<MysqlFormValues>({
    resolver: yupResolver(MYSQL_FORM_SCHEMA),
    values: defaultValues,
    context: {
      originalConnectionName: defaultValues.connectionName,
      accountId: account?.id ?? '',
    },
  });
  const [validationResponse, setValidationResponse] = useState<
    CheckConnectionConfigResponse | undefined
  >();

  const [isValidating, setIsValidating] = useState<boolean>(false);

  const [openPermissionDialog, setOpenPermissionDialog] =
    useState<boolean>(false);
  const [permissionData, setPermissionData] =
    useState<ConnectionRolePrivilege[]>();
  async function onSubmit(values: MysqlFormValues) {
    try {
      const connectionResp = await updateMysqlConnection(
        connectionId,
        values.connectionName,
        values.db,
        values.tunnel,
        account?.id ?? '',
        values.options
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
                <Input type="password" placeholder="mysql" {...field} />
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
        <FormField
          control={form.control}
          name="options.maxConnectionLimit"
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Max Connection Limit</FormLabel>
                <FormDescription>
                  The maximum number of concurrent database connections allowed.
                  If set to 0 then there is no limit on the number of open
                  connections.
                </FormDescription>
              </div>
              <FormControl>
                <Input
                  {...field}
                  className="max-w-[180px]"
                  type="number"
                  value={field.value ? field.value.toString() : 0}
                  onChange={(event) => {
                    field.onChange(event.target.valueAsNumber);
                  }}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Accordion type="single" collapsible className="w-full">
          <AccordionItem value="bastion">
            <AccordionTrigger> Bastion Host Configuration</AccordionTrigger>
            <AccordionContent className="flex flex-col gap-4 p-2">
              <div className="text-sm">
                This section is optional and only necessary if your database is
                not publicly accessible to the internet.
              </div>
              <FormField
                control={form.control}
                name="tunnel.host"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Host</FormLabel>
                    <FormDescription>
                      The hostname of the bastion server that will be used for
                      SSH tunneling.
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
                      The private key that will be used to authenticate against
                      the SSH server. If using passphrase auth, provide that in
                      the appropriate field below.
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
                      The passphrase that will be used to authenticate with. If
                      the SSH Key provided above is encrypted, provide the
                      password for it here.
                    </FormDescription>
                    <FormControl>
                      <Input type="password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="tunnel.knownHostPublicKey"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Known Host Public Key</FormLabel>
                    <FormDescription>
                      The public key of the host that will be expected when
                      connecting to the tunnel. This should be in the format
                      like what is found in the `~/.ssh/known_hosts` file,
                      excluding the hostname. If this is not provided, any host
                      public key will be accepted. Currently only a single host
                      key is supported.
                    </FormDescription>
                    <FormControl>
                      <Input
                        placeholder="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAlkjd9s7aJkfdLk3jSLkfj2lk3j2lkfj2l3kjf2lkfj2l"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </AccordionContent>
          </AccordionItem>
        </Accordion>
        <div className="flex flex-row gap-3 justify-between">
          <Button
            variant="outline"
            disabled={!form.formState.isValid}
            onClick={async () => {
              setIsValidating(true);
              try {
                const res = await checkMysqlConnection(
                  form.getValues('db'),
                  form.getValues('tunnel'),
                  account?.id ?? ''
                );
                setIsValidating(false);
                setValidationResponse(res);
                setPermissionData(res.privileges);
                setOpenPermissionDialog(res?.isConnected && true);
              } catch (err) {
                setIsValidating(false);
                setValidationResponse(
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
            <ButtonText
              leftIcon={
                isValidating ? (
                  <Spinner className="text-black dark:text-white" />
                ) : (
                  <div></div>
                )
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
        {validationResponse && validationResponse.isConnected && (
          <SuccessAlert description={'Successfully connected!'} />
        )}
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

interface SuccessAlertProps {
  description: string;
}

function SuccessAlert(props: SuccessAlertProps): ReactElement {
  const { description } = props;
  return (
    <Alert variant="success">
      <div className="flex flex-row items-center gap-2">
        <CheckCircledIcon className="h-4 w-4 text-green-900 dark:text-green-400" />
        <div className="font-normal text-green-900 dark:text-green-400">
          {description}
        </div>
      </div>
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
  tunnel: MysqlFormValues['tunnel'],
  accountId: string,
  options: MysqlFormValues['options']
): Promise<UpdateConnectionResponse> {
  const myconfig = new MysqlConnectionConfig({
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
  });
  if (options && options.maxConnectionLimit != 0) {
    myconfig.connectionOptions = new SqlConnectionOptions({
      maxConnectionLimit: options.maxConnectionLimit,
    });
  }
  if (tunnel && tunnel.host) {
    myconfig.tunnel = new SSHTunnel({
      host: tunnel.host,
      port: tunnel.port,
      user: tunnel.user,
      knownHostPublicKey: tunnel.knownHostPublicKey
        ? tunnel.knownHostPublicKey
        : undefined,
    });
    if (tunnel.privateKey) {
      myconfig.tunnel.authentication = new SSHAuthentication({
        authConfig: {
          case: 'privateKey',
          value: new SSHPrivateKey({
            value: tunnel.privateKey,
            passphrase: tunnel.passphrase,
          }),
        },
      });
    } else if (tunnel.passphrase) {
      myconfig.tunnel.authentication = new SSHAuthentication({
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
              case: 'mysqlConfig',
              value: myconfig,
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
