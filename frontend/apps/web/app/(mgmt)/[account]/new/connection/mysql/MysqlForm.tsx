'use client';
import ButtonText from '@/components/ButtonText';
import { PasswordInput } from '@/components/PasswordComponent';
import Spinner from '@/components/Spinner';
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
  MYSQL_CONNECTION_PROTOCOLS,
  MysqlCreateConnectionFormContext,
  MysqlFormValues,
} from '@/yup-validations/connections';
import { create } from '@bufbuild/protobuf';
import { createConnectQueryKey, useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckConnectionConfigResponse,
  CheckConnectionConfigResponseSchema,
  GetConnectionResponseSchema,
} from '@neosync/sdk';
import {
  checkConnectionConfig,
  createConnection,
  getConnection,
  isConnectionNameAvailable,
} from '@neosync/sdk/connectquery';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useQueryClient } from '@tanstack/react-query';
import { useRouter, useSearchParams } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { buildConnectionConfigMysql } from '../../../connections/util';

type ActiveTab = 'host' | 'url';

export default function MysqlForm() {
  const { account } = useAccount();
  const searchParams = useSearchParams();
  const sourceConnId = searchParams.get('sourceId');
  const [isLoading, setIsLoading] = useState<boolean>();

  // used to know which tab - host or url that the user is on when we submit the form
  const [activeTab, setActiveTab] = useState<ActiveTab>('url');
  const { mutateAsync: isConnectionNameAvailableAsync } = useMutation(
    isConnectionNameAvailable
  );
  const form = useForm<MysqlFormValues, MysqlCreateConnectionFormContext>({
    resolver: yupResolver(MysqlFormValues),
    defaultValues: {
      connectionName: '',
      db: {
        host: 'localhost',
        name: 'mysql',
        user: 'mysql',
        pass: 'mysql',
        port: 3306,
        protocol: 'tcp',
      },
      url: '',
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
        clientCert: '',
        clientKey: '',
        rootCert: '',
        serverName: '',
      },
    },
    context: {
      accountId: account?.id ?? '',
      activeTab: activeTab,
      isConnectionNameAvailable: isConnectionNameAvailableAsync,
    },
  });
  const router = useRouter();
  const [validationResponse, setValidationResponse] = useState<
    CheckConnectionConfigResponse | undefined
  >();

  const [isValidating, setIsValidating] = useState<boolean>(false);
  const [openPermissionDialog, setOpenPermissionDialog] =
    useState<boolean>(false);
  const posthog = usePostHog();
  const { mutateAsync: createMysqlConnection } = useMutation(createConnection);
  const { mutateAsync: checkMysqlConnection } = useMutation(
    checkConnectionConfig
  );
  const { mutateAsync: getMysqlConnection } = useMutation(getConnection);
  const queryclient = useQueryClient();
  async function onSubmit(values: MysqlFormValues) {
    if (!account) {
      return;
    }
    try {
      const connection = await createMysqlConnection({
        name: values.connectionName,
        accountId: account.id,
        connectionConfig: buildConnectionConfigMysql({
          ...values,
          url: activeTab === 'url' ? values.url : undefined,
          db: values.db,
        }),
      });
      posthog.capture('New Connection Created', { type: 'mysql' });
      toast.success('Successfully created connection!');
      const returnTo = searchParams.get('returnTo');
      if (returnTo) {
        router.push(returnTo);
      } else if (connection.connection?.id) {
        queryclient.setQueryData(
          createConnectQueryKey({
            schema: getConnection,
            input: { id: connection.connection.id },
            cardinality: undefined,
          }),
          create(GetConnectionResponseSchema, {
            connection: connection.connection,
          })
        );
        router.push(
          `/${account?.name}/connections/${connection.connection.id}`
        );
      } else {
        router.push(`/${account?.name}/connections`);
      }
    } catch (err) {
      console.error(err);
      toast.error('Unable to create connection', {
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
        const connData = await getMysqlConnection({ id: sourceConnId });
        if (
          connData.connection?.connectionConfig?.config.case !== 'mysqlConfig'
        ) {
          return;
        }

        const config = connData.connection?.connectionConfig?.config.value;
        const mysqlConfig = config.connectionConfig.value;

        const dbConfig = {
          host: '',
          name: '',
          user: '',
          pass: '',
          port: 3306,
          protocol: 'tcp',
        };
        if (typeof mysqlConfig !== 'string') {
          dbConfig.host = mysqlConfig?.host ?? '';
          dbConfig.name = mysqlConfig?.name ?? '';
          dbConfig.user = mysqlConfig?.user ?? '';
          dbConfig.pass = mysqlConfig?.pass ?? '';
          dbConfig.port = mysqlConfig?.port ?? 3306;
          dbConfig.protocol = mysqlConfig?.protocol ?? 'tcp';
        }

        let passPhrase = '';
        let privateKey = '';

        const authConfig = config.tunnel?.authentication?.authConfig;

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
          db: dbConfig,
          url: typeof mysqlConfig === 'string' ? mysqlConfig : '',
          options: {
            maxConnectionLimit:
              config.connectionOptions?.maxConnectionLimit ?? 50,
            maxIdleDuration: config.connectionOptions?.maxIdleDuration ?? '',
            maxIdleLimit: config.connectionOptions?.maxIdleConnections ?? 2,
            maxOpenDuration: config.connectionOptions?.maxOpenDuration ?? '',
          },
          tunnel: {
            host: config.tunnel?.host ?? '',
            port: config.tunnel?.port ?? 22,
            knownHostPublicKey: config.tunnel?.knownHostPublicKey ?? '',
            user: config.tunnel?.user ?? '',
            passphrase: passPhrase,
            privateKey: privateKey,
          },
          clientTls: {
            clientCert: config.clientTls?.clientCert ?? '',
            clientKey: config.clientTls?.clientKey ?? '',
            rootCert: config.clientTls?.rootCert ?? '',
            serverName: config.clientTls?.serverName ?? '',
          },
        });
        if (config.connectionConfig.case === 'url') {
          setActiveTab('url');
        } else if (config.connectionConfig.case === 'connection') {
          setActiveTab('host');
        }
      } catch (error) {
        console.error('Failed to fetch connection data:', error);
        setIsLoading(false);
        toast.error('Unable to retrieve connection data from clone!', {
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
          defaultValue="url"
          onValueChange={(value) => setActiveTab(value as ActiveTab)}
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
          </div>
        </RadioGroup>
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
                    placeholder="username:password@tcp(hostname:port)/database"
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
                  <FormDescription>The database port</FormDescription>
                  <FormControl>
                    <Input
                      type="number"
                      placeholder="3306"
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
                    <Input placeholder="mysql" {...field} />
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
                    <PasswordInput placeholder="mysql" {...field} />
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
                    The protocol that you want to use to connect to your
                    database
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
          connectionType="mysql"
        />
        <div className="flex flex-row gap-3 justify-between">
          <Button
            variant="outline"
            disabled={!form.formState.isValid}
            onClick={async () => {
              setIsValidating(true);
              const values = form.getValues();
              try {
                const res = await checkMysqlConnection({
                  connectionConfig: buildConnectionConfigMysql({
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
