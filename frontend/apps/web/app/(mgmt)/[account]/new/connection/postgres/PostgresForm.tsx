'use client';
import ButtonText from '@/components/ButtonText';
import FormError from '@/components/FormError';
import Spinner from '@/components/Spinner';
import RequiredLabel from '@/components/labels/RequiredLabel';
import { setOnboardingConfig } from '@/components/onboarding-checklist/OnboardingChecklist';
import Permissions from '@/components/permissions/Permissions';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
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
import { toast } from '@/components/ui/use-toast';
import { useGetAccountOnboardingConfig } from '@/libs/hooks/useGetAccountOnboardingConfig';
import { getErrorMessage } from '@/util/util';
import {
  POSTGRES_FORM_SCHEMA,
  PostgresFormValues,
  SSL_MODES,
} from '@/yup-validations/connections';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckConnectionConfigResponse,
  ConnectionConfig,
  ConnectionRolePrivilege,
  CreateConnectionRequest,
  CreateConnectionResponse,
  GetConnectionResponse,
  IsConnectionNameAvailableResponse,
  PostgresConnection,
  PostgresConnectionConfig,
  SSHAuthentication,
  SSHPassphrase,
  SSHPrivateKey,
  SSHTunnel,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useRouter, useSearchParams } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';

export default function PostgresForm() {
  const searchParams = useSearchParams();
  const { account } = useAccount();
  const sourceConnId = searchParams.get('sourceId');
  const [isLoading, setIsLoading] = useState<boolean>();
  const { data: onboardingData, mutate } = useGetAccountOnboardingConfig(
    account?.id ?? ''
  );
  // used to know which tab - host or url that the user is on when we submit the form
  const [activeTab, setActiveTab] = useState<string>('url');

  const form = useForm<PostgresFormValues>({
    resolver: yupResolver(POSTGRES_FORM_SCHEMA),
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
      tunnel: {
        host: '',
        port: 22,
        knownHostPublicKey: '',
        user: '',
        passphrase: '',
        privateKey: '',
      },
    },

    context: { accountId: account?.id ?? '', activeTab: activeTab },
  });
  const router = useRouter();
  const [validationResponse, setValidationResponse] = useState<
    CheckConnectionConfigResponse | undefined
  >();

  const [isValidating, setIsValidating] = useState<boolean>(false);

  const [openPermissionDialog, setOpenPermissionDialog] =
    useState<boolean>(false);
  const [permissionData, setPermissionData] =
    useState<ConnectionRolePrivilege[]>();

  async function onSubmit(values: PostgresFormValues) {
    if (!account) {
      return;
    }

    try {
      let connection: CreateConnectionResponse = new CreateConnectionResponse(
        {}
      );
      if (activeTab === 'host') {
        connection = await createPostgresConnection(
          values.connectionName,
          account.id,
          values.db,
          undefined, // don't pass in the url since user is submitting the db values
          values.tunnel
        );
      } else if (activeTab === 'url') {
        connection = await createPostgresConnection(
          values.connectionName,
          account.id,
          undefined, // don't pass in the db values since user is submitting the url
          values.url ?? '',
          values.tunnel
        );
      }

      toast({
        title: 'Successfully created connection!',
        variant: 'success',
      });

      // updates the onboarding data
      if (onboardingData?.config?.hasCreatedSourceConnection) {
        try {
          await setOnboardingConfig(account.id, {
            hasCreatedSourceConnection:
              onboardingData.config.hasCreatedSourceConnection,
            hasCreatedDestinationConnection: true,
            hasCreatedJob: onboardingData.config.hasCreatedJob,
            hasInvitedMembers: onboardingData.config.hasInvitedMembers,
          });
          mutate();
        } catch (e) {
          toast({
            title: 'Unable to update onboarding status!',
            variant: 'destructive',
          });
        }
      } else {
        try {
          await setOnboardingConfig(account.id, {
            hasCreatedSourceConnection: true,
            hasCreatedDestinationConnection:
              onboardingData?.config?.hasCreatedSourceConnection ?? true,
            hasCreatedJob: onboardingData?.config?.hasCreatedJob ?? true,
            hasInvitedMembers:
              onboardingData?.config?.hasInvitedMembers ?? true,
          });
          mutate();
        } catch (e) {
          toast({
            title: 'Unable to update onboarding status!',
            variant: 'destructive',
          });
        }
      }

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
      toast({
        title: 'Unable to create connection',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  /* we call the underlying useGetConnection API directly since we can't call
the hook in the useEffect conditionally. This is used to retrieve the values for the clone connection so that we can update the form. 
*/
  useEffect(() => {
    const fetchData = async () => {
      if (sourceConnId && account?.id) {
        setIsLoading(true);
        try {
          const connData = await GetConnectionCloneValues(
            account.id,
            sourceConnId
          );

          if (
            connData &&
            connData.connection?.connectionConfig?.config.case === 'pgConfig'
          ) {
            const config = connData.connection?.connectionConfig?.config.value;

            let pgConfig: PostgresConnection | string | undefined;

            if (config.connectionConfig.case == 'connection') {
              pgConfig = config.connectionConfig.value;
            } else if (config.connectionConfig.case == 'url') {
              pgConfig = config.connectionConfig.value;
            }

            const defaultDb = {
              host: '',
              name: '',
              user: '',
              pass: '',
              port: 5432,
              sslMode: 'disable',
            };

            let dbConfig = defaultDb;
            if (typeof pgConfig !== 'string') {
              dbConfig = {
                host: pgConfig?.host ?? '',
                name: pgConfig?.name ?? '',
                user: pgConfig?.user ?? '',
                pass: pgConfig?.pass ?? '',
                port: pgConfig?.port ?? 5432,
                sslMode: pgConfig?.sslMode ?? 'disable',
              };
            }

            /* reset the form with the new values and include the fallback values because of our validation schema requires a string and not undefined which is okay because it will tell the user that something is wrong instead of the user not realizing that it's undefined
             */
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

            form.reset({
              ...form.getValues(),
              connectionName: connData.connection?.name + '-copy',
              db: dbConfig,
              url: typeof pgConfig === 'string' ? pgConfig : '',
              tunnel: {
                host: config.tunnel?.host ?? '',
                port: config.tunnel?.port ?? 22,
                knownHostPublicKey: config.tunnel?.knownHostPublicKey ?? '',
                user: config.tunnel?.user ?? '',
                passphrase: passPhrase,
                privateKey: privateKey,
              },
            });
          }
        } catch (error) {
          console.error('Failed to fetch connection data:', error);
          setIsLoading(false);
          toast({
            title: 'Unable to clone connection!',
            variant: 'destructive',
          });
        } finally {
          setIsLoading(false);
        }
      }
    };

    fetchData();
  }, [account?.id]);

  if (isLoading || !account?.id) {
    return <SkeletonTable />;
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

        <RadioGroup
          defaultValue="url"
          onValueChange={(e) => setActiveTab(e)}
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
                    <Input type="password" placeholder="postgres" {...field} />
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
        <Permissions
          data={permissionData ?? []}
          openPermissionDialog={openPermissionDialog}
          setOpenPermissionDialog={setOpenPermissionDialog}
          isValidating={isValidating}
          validationResponse={validationResponse?.isConnected ?? false}
          connectionName={form.getValues('connectionName')}
        />
        <div className="flex flex-row gap-3 justify-between">
          <Button
            variant="outline"
            onClick={async () => {
              setIsValidating(true);
              try {
                let res: CheckConnectionConfigResponse =
                  new CheckConnectionConfigResponse({});
                if (activeTab === 'host') {
                  res = await checkPostgresConnection(
                    account?.id ?? '',
                    form.getValues().db,
                    form.getValues().tunnel,
                    undefined
                  );
                } else if (activeTab === 'url') {
                  res = await checkPostgresConnection(
                    account?.id ?? '',
                    undefined,
                    form.getValues().tunnel,
                    form.getValues().url ?? ''
                  );
                }
                setIsValidating(false);
                setValidationResponse(res);
                setPermissionData(res.privileges);
                setOpenPermissionDialog(res?.isConnected && true);
              } catch (err) {
                setValidationResponse(
                  new CheckConnectionConfigResponse({
                    isConnected: false,
                    connectionError:
                      err instanceof Error ? err.message : 'unknown error',
                  })
                );
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
async function createPostgresConnection(
  name: string,
  accountId: string,
  db?: PostgresFormValues['db'],
  url?: string,
  tunnel?: PostgresFormValues['tunnel']
): Promise<CreateConnectionResponse> {
  let pgconfig = new PostgresConnectionConfig({});

  if (url) {
    pgconfig.connectionConfig = {
      case: 'url',
      value: url,
    };
  } else {
    pgconfig.connectionConfig = {
      case: 'connection',
      value: new PostgresConnection({
        host: db?.host,
        name: db?.name,
        user: db?.user,
        pass: db?.pass,
        port: db?.port,
        sslMode: db?.sslMode,
      }),
    };
  }
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
  const res = await fetch(`/api/accounts/${accountId}/connections`, {
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
            value: pgconfig,
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
  accountId: string,
  db?: PostgresFormValues['db'],
  tunnel?: PostgresFormValues['tunnel'],
  url?: string
): Promise<CheckConnectionConfigResponse> {
  let requestBody;
  if (url) {
    requestBody = { url, tunnel };
  } else {
    requestBody = { db, tunnel };
  }
  const res = await fetch(
    `/api/accounts/${accountId}/connections/postgres/check`,
    {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(requestBody),
    }
  );
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
    `/api/accounts/${accountId}/connections/is-connection-name-available?connectionName=${name}`,
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

export async function GetConnectionCloneValues(
  accountId: string,
  sourceConnId: string
): Promise<GetConnectionResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/connections/${sourceConnId}`,
    {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    }
  );

  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }

  return GetConnectionResponse.fromJson(await res.json());
}
