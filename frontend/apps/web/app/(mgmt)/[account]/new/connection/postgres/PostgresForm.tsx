'use client';
import ButtonText from '@/components/ButtonText';
import { PasswordInput } from '@/components/PasswordComponent';
import Spinner from '@/components/Spinner';
import OSSOnlyGuard from '@/components/guards/OSSOnlyGuard';
import RequiredLabel from '@/components/labels/RequiredLabel';
import PermissionsDialog from '@/components/permissions/PermissionsDialog';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
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
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { getErrorMessage } from '@/util/util';
import {
  ActiveConnectionTab,
  PostgresCreateConnectionFormContext,
  PostgresFormValues,
  SSL_MODES,
} from '@/yup-validations/connections';
import { create } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckConnectionConfigResponse,
  CheckConnectionConfigResponseSchema,
  ConnectionService,
  PostgresConnectionConfig,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useRouter, useSearchParams } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { buildConnectionConfigPostgres } from '../../../connections/util';

export default function PostgresForm() {
  const searchParams = useSearchParams();
  const { account } = useAccount();
  const sourceConnId = searchParams.get('sourceId');
  const [isLoading, setIsLoading] = useState<boolean>();
  const { mutateAsync: isConnectionNameAvailableAsync } = useMutation(
    ConnectionService.method.isConnectionNameAvailable
  );

  // used to know which tab - host or url that the user is on when we submit the form
  const [activeTab, setActiveTab] = useState<ActiveConnectionTab>('url');

  const form = useForm<PostgresFormValues, PostgresCreateConnectionFormContext>(
    {
      resolver: yupResolver(PostgresFormValues),
      mode: 'onChange',
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
        url: '',
        envVar: '',
        options: {
          maxConnectionLimit: 50,
          maxIdleDuration: '',
          maxIdleLimit: 2,
          maxOpenDuration: '',
        },
        tunnel: {
          host: '',
          port: 22,
          knownHostPublicKey: '',
          user: '',
          passphrase: '',
          privateKey: '',
        },
        clientTls: {
          rootCert: '',
          clientCert: '',
          clientKey: '',
          serverName: '',
        },
      },
      context: {
        accountId: account?.id ?? '',
        activeTab: activeTab,
        isConnectionNameAvailable: isConnectionNameAvailableAsync,
      },
    }
  );
  const { mutateAsync: createPostgresConnection } = useMutation(
    ConnectionService.method.createConnection
  );
  const { mutateAsync: checkPostgresConnection } = useMutation(
    ConnectionService.method.checkConnectionConfig
  );
  const { mutateAsync: getPostgresConnection } = useMutation(
    ConnectionService.method.getConnection
  );

  const router = useRouter();
  const [validationResponse, setValidationResponse] = useState<
    CheckConnectionConfigResponse | undefined
  >();

  const [isValidating, setIsValidating] = useState<boolean>(false);
  const [openPermissionDialog, setOpenPermissionDialog] =
    useState<boolean>(false);
  const posthog = usePostHog();

  async function onSubmit(values: PostgresFormValues) {
    if (!account) {
      return;
    }

    try {
      const connection = await createPostgresConnection({
        name: values.connectionName,
        accountId: account.id,
        connectionConfig: buildConnectionConfigPostgres({
          ...values,
          url: activeTab === 'url' ? values.url : undefined,
          db: activeTab === 'host' ? values.db : {},
          envVar: activeTab === 'url-env' ? values.envVar : undefined,
        }),
      });
      posthog.capture('New Connection Created', { type: 'postgres' });
      toast.success('Successfully created connection!');
      const returnTo = searchParams.get('returnTo');
      if (returnTo) {
        router.push(returnTo);
      } else if (connection.connection?.id) {
        router.push(
          `/${account?.name}/connections/${connection.connection.id}`
        );
      } else {
        router.push(`/${account?.name}/connections`);
      }
    } catch (err) {
      console.error('Error in form submission:', err);
      toast.error('Unable to create connection!', {
        description: getErrorMessage(err),
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
        const connData = await getPostgresConnection({ id: sourceConnId });
        const connectionConfig = connData.connection?.connectionConfig?.config;
        if (!connectionConfig || connectionConfig.case !== 'pgConfig') {
          return;
        }

        const { db, url, envVar } = getPgConnectionFormValues(
          connectionConfig.value
        );

        let passPhrase = '';
        let privateKey = '';

        const authConfig =
          connectionConfig.value.tunnel?.authentication?.authConfig;
        switch (authConfig?.case) {
          case 'passphrase':
            passPhrase = authConfig.value.value;
            break;
          case 'privateKey':
            passPhrase = authConfig.value.passphrase ?? '';
            privateKey = authConfig.value.value;
            break;
        }

        /* reset the form with the new values and include the fallback values because of our validation schema requires a string and not undefined which is okay because it will tell the user that something is wrong instead of the user not realizing that it's undefined
         */
        form.reset({
          ...form.getValues(),
          connectionName: connData.connection?.name + '-copy',
          db,
          url,
          envVar,
          options: {
            maxConnectionLimit:
              connectionConfig.value.connectionOptions?.maxConnectionLimit ??
              50,
            maxIdleDuration:
              connectionConfig.value.connectionOptions?.maxIdleDuration ?? '',
            maxIdleLimit:
              connectionConfig.value.connectionOptions?.maxIdleConnections ?? 2,
            maxOpenDuration:
              connectionConfig.value.connectionOptions?.maxOpenDuration ?? '',
          },
          tunnel: {
            host: connectionConfig.value.tunnel?.host ?? '',
            port: connectionConfig.value.tunnel?.port ?? 22,
            knownHostPublicKey:
              connectionConfig.value.tunnel?.knownHostPublicKey ?? '',
            user: connectionConfig.value.tunnel?.user ?? '',
            passphrase: passPhrase,
            privateKey: privateKey,
          },
          clientTls: {
            clientCert: connectionConfig.value.clientTls?.clientCert ?? '',
            clientKey: connectionConfig.value.clientTls?.clientKey ?? '',
            rootCert: connectionConfig.value.clientTls?.rootCert ?? '',
            serverName: connectionConfig.value.clientTls?.serverName ?? '',
          },
        });
        if (connectionConfig.value.connectionConfig.case === 'url') {
          setActiveTab('url');
        } else if (
          connectionConfig.value.connectionConfig.case === 'connection'
        ) {
          setActiveTab('host');
        } else if (
          connectionConfig.value.connectionConfig.case === 'urlFromEnv'
        ) {
          setActiveTab('url-env');
        }
      } catch (error) {
        console.error('Failed to fetch connection data:', error);
        toast.error('Unable to retrieve connection data for clone!', {
          description: getErrorMessage(error),
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
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
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
              <FormMessage />
            </FormItem>
          )}
        />

        <RadioGroup
          defaultValue={activeTab}
          onValueChange={(e) => setActiveTab(e as ActiveConnectionTab)}
          value={activeTab}
        >
          <div className="flex flex-col md:flex-row gap-4">
            <div className="text-sm">Connect by:</div>
            <div className="flex items-center space-x-2">
              <RadioGroupItem value="url" id="r2" />
              <Label htmlFor="r2">URL</Label>
            </div>
            <div className="flex items-center space-x-2">
              <RadioGroupItem value="host" id="r1" />
              <Label htmlFor="r1">Host</Label>
            </div>
            <OSSOnlyGuard>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="url-env" id="r3" />
                <Label htmlFor="r3">Environment Variable</Label>
              </div>
            </OSSOnlyGuard>
          </div>
        </RadioGroup>

        {activeTab === 'url-env' && (
          <FormField
            control={form.control}
            name="envVar"
            render={({ field }) => (
              <FormItem>
                <FormLabel>
                  <RequiredLabel />
                  Environment Variable
                </FormLabel>
                <FormDescription>
                  The environment variable that contains the connection URL.
                  Must start with &quot;USER_DEFINED_&quot;. Must be present on
                  both the backend and the worker processes for full
                  functionality.
                </FormDescription>
                <FormControl>
                  <Input placeholder="USER_DEFINED_POSTGRES_URL" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        )}

        {activeTab == 'url' && (
          <FormField
            control={form.control}
            name="url"
            render={({ field }) => (
              <FormItem>
                <FormLabel>
                  <RequiredLabel />
                  Connection URL
                </FormLabel>
                <FormDescription>Your connection URL</FormDescription>
                <FormControl>
                  <Input
                    placeholder="postgres://test:test@host.com?sslMode=require"
                    {...field}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        )}
        {activeTab == 'host' && (
          <>
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
                    <Input
                      type="number"
                      placeholder="5432"
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
                    <PasswordInput placeholder="postgres" {...field} />
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
          </>
        )}
        <div className="flex flex-col gap-0">
          <FormField
            control={form.control}
            name="options.maxConnectionLimit"
            render={({ field }) => (
              <FormItem>
                <div className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                  <div className="space-y-0.5">
                    <FormLabel>Max Open Connection Limit</FormLabel>
                    <FormDescription>
                      The maximum number of concurrent database connections
                      allowed. If set to 0 then there is no limit on the number
                      of open connections. -1 to leave unset and use system
                      default.
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Input
                      {...field}
                      className="max-w-[180px]"
                      type="number"
                      value={
                        field.value != null ? field.value.toString() : '-1'
                      }
                      onChange={(event) => {
                        field.onChange(event.target.valueAsNumber);
                      }}
                    />
                  </FormControl>
                </div>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="options.maxOpenDuration"
            render={({ field }) => (
              <FormItem>
                <div className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                  <div className="space-y-0.5">
                    <FormLabel>Max Open Duration</FormLabel>
                    <FormDescription>
                      The maximum amount of time a connection may be reused.
                      Expired connections may be closed laizly before reuse. Ex:
                      1s, 1m, 500ms. Empty to leave unset.
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Input
                      {...field}
                      className="max-w-[180px]"
                      value={field.value ? field.value : ''}
                      onChange={(event) => {
                        field.onChange(event.target.value);
                      }}
                    />
                  </FormControl>
                </div>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="options.maxIdleLimit"
            render={({ field }) => (
              <FormItem>
                <div className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                  <div className="space-y-0.5">
                    <FormLabel>Max Idle Connection Limit</FormLabel>
                    <FormDescription>
                      The maximum number of idle database connections allowed.
                      If set to 0 then there is no limit on the number of idle
                      connections. -1 to leave unset and use system default.
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Input
                      {...field}
                      className="max-w-[180px]"
                      type="number"
                      value={
                        field.value != null ? field.value.toString() : '-1'
                      }
                      onChange={(event) => {
                        field.onChange(event.target.valueAsNumber);
                      }}
                    />
                  </FormControl>
                </div>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="options.maxIdleDuration"
            render={({ field }) => (
              <FormItem>
                <div className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                  <div className="space-y-0.5">
                    <FormLabel>Max Idle Duration</FormLabel>
                    <FormDescription>
                      The maximum amount of time a connection may be idle.
                      Expired connections may be closed laizly before reuse. Ex:
                      1s, 1m, 500ms. Empty to leave unset.
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Input
                      {...field}
                      className="max-w-[180px]"
                      value={field.value ? field.value : ''}
                      onChange={(event) => {
                        field.onChange(event.target.value);
                      }}
                    />
                  </FormControl>
                </div>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
        <Accordion type="single" collapsible className="w-full">
          <AccordionItem value="bastion">
            <AccordionTrigger>Client TLS Certificates</AccordionTrigger>
            <AccordionContent className="flex flex-col gap-4 p-2">
              <div className="text-sm">
                Configuring this section allows Neosync to connect to the
                database using SSL/TLS. The verification mode may be configured
                using the SSL Field, or by specifying the option in the
                postgresql url.
              </div>
              <FormField
                control={form.control}
                name="clientTls.rootCert"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Root Certificate</FormLabel>
                    <FormDescription>
                      {`The public key certificate of the CA that issued the
                      server's certificate. Root certificates are used to
                      authenticate the server to the client. They ensure that
                      the server the client is connecting to is trusted.`}
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
                name="clientTls.clientCert"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Client Certificate</FormLabel>
                    <FormDescription>
                      A public key certificate issued to the client by a trusted
                      Certificate Authority (CA).
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
                name="clientTls.clientKey"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Client Key</FormLabel>
                    <FormDescription>
                      A private key corresponding to the client certificate.
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
                name="clientTls.serverName"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Server Name</FormLabel>
                    <FormDescription>
                      {`Server Name is used to verify the hostname on the returned
                      certificates. It is also included in the client's
                      handshake to support virtual hosting unless it is an IP
                      address. This is only required if performing full tls
                      verification.`}
                    </FormDescription>
                    <FormControl>
                      <Textarea {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </AccordionContent>
          </AccordionItem>
        </Accordion>

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
                      public key will be accepted.
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
        <PermissionsDialog
          checkResponse={
            validationResponse ??
            create(CheckConnectionConfigResponseSchema, {})
          }
          openPermissionDialog={openPermissionDialog}
          setOpenPermissionDialog={setOpenPermissionDialog}
          isValidating={isValidating}
          connectionName={form.getValues('connectionName')}
          connectionType="postgres"
        />
        <div className="flex flex-row gap-3 justify-between">
          <Button
            variant="outline"
            onClick={async () => {
              setIsValidating(true);
              try {
                const values = form.getValues();
                const res = await checkPostgresConnection({
                  connectionConfig: buildConnectionConfigPostgres({
                    ...values,
                    url: activeTab === 'url' ? values.url : undefined,
                    db: values.db,
                  }),
                });
                setValidationResponse(res);
                setOpenPermissionDialog(!!res?.isConnected);
              } catch (err) {
                setValidationResponse(
                  create(CheckConnectionConfigResponseSchema, {
                    isConnected: false,
                    connectionError:
                      err instanceof Error ? err.message : 'unknown error',
                  })
                );
              } finally {
                setIsValidating(false);
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

// extracts the connection config and returns the values for the form
export function getPgConnectionFormValues(
  connection: PostgresConnectionConfig
): Pick<PostgresFormValues, 'db' | 'url' | 'envVar'> {
  switch (connection.connectionConfig.case) {
    case 'connection':
      return {
        db: connection.connectionConfig.value,
        url: undefined,
        envVar: undefined,
      };
    case 'url':
      return {
        db: {},
        url: connection.connectionConfig.value,
        envVar: undefined,
      };
    case 'urlFromEnv':
      return {
        db: {},
        url: undefined,
        envVar: connection.connectionConfig.value,
      };
    default:
      return {
        db: {},
        url: undefined,
        envVar: undefined,
      };
  }
}
