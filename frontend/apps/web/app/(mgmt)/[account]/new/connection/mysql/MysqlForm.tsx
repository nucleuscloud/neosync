'use client';
import ButtonText from '@/components/ButtonText';
import FormError from '@/components/FormError';
import { PasswordInput } from '@/components/PasswordComponent';
import Spinner from '@/components/Spinner';
import RequiredLabel from '@/components/labels/RequiredLabel';
import { buildAccountOnboardingConfig } from '@/components/onboarding-checklist/OnboardingChecklist';
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
import { toast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
import {
  MYSQL_CONNECTION_PROTOCOLS,
  MysqlFormValues,
} from '@/yup-validations/connections';
import {
  createConnectQueryKey,
  useMutation,
  useQuery,
} from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckConnectionConfigResponse,
  GetAccountOnboardingConfigResponse,
  GetConnectionResponse,
} from '@neosync/sdk';
import {
  checkConnectionConfig,
  createConnection,
  getAccountOnboardingConfig,
  getConnection,
  setAccountOnboardingConfig,
} from '@neosync/sdk/connectquery';
import {
  CheckCircledIcon,
  ExclamationTriangleIcon,
} from '@radix-ui/react-icons';
import { useQueryClient } from '@tanstack/react-query';
import { useRouter, useSearchParams } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { buildConnectionConfigMysql } from '../../../connections/util';

type ActiveTab = 'host' | 'url';

export default function MysqlForm() {
  const { account } = useAccount();
  const searchParams = useSearchParams();
  const sourceConnId = searchParams.get('sourceId');
  const [isLoading, setIsLoading] = useState<boolean>();

  // used to know which tab - host or url that the user is on when we submit the form
  const [activeTab, setActiveTab] = useState<ActiveTab>('url');

  const form = useForm<MysqlFormValues>({
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
        maxConnectionLimit: 80,
      },
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
  const posthog = usePostHog();
  const { mutateAsync: createMysqlConnection } = useMutation(createConnection);
  const { mutateAsync: checkMysqlConnection } = useMutation(
    checkConnectionConfig
  );
  const { mutateAsync: getMysqlConnection } = useMutation(getConnection);
  const { data: onboardingData } = useQuery(
    getAccountOnboardingConfig,
    { accountId: account?.id ?? '' },
    { enabled: !!account?.id }
  );
  const queryclient = useQueryClient();
  const { mutateAsync: setOnboardingConfigAsync } = useMutation(
    setAccountOnboardingConfig
  );

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
      toast({
        title: 'Successfully created connection!',
        variant: 'success',
      });

      // updates the onboarding data
      if (onboardingData?.config?.hasCreatedSourceConnection) {
        try {
          const resp = await setOnboardingConfigAsync({
            accountId: account.id,
            config: buildAccountOnboardingConfig({
              hasCreatedSourceConnection:
                onboardingData.config.hasCreatedSourceConnection,
              hasCreatedDestinationConnection: true,
              hasCreatedJob: onboardingData.config.hasCreatedJob,
              hasInvitedMembers: onboardingData.config.hasInvitedMembers,
            }),
          });
          queryclient.setQueryData(
            createConnectQueryKey(getAccountOnboardingConfig, {
              accountId: account.id,
            }),
            new GetAccountOnboardingConfigResponse({
              config: resp.config,
            })
          );
        } catch (e) {
          toast({
            title: 'Unable to update onboarding status!',
            variant: 'destructive',
          });
        }
      } else {
        try {
          const resp = await setOnboardingConfigAsync({
            accountId: account.id,
            config: buildAccountOnboardingConfig({
              hasCreatedSourceConnection: true,
              hasCreatedDestinationConnection:
                onboardingData?.config?.hasCreatedSourceConnection ?? true,
              hasCreatedJob: onboardingData?.config?.hasCreatedJob ?? true,
              hasInvitedMembers:
                onboardingData?.config?.hasInvitedMembers ?? true,
            }),
          });
          queryclient.setQueryData(
            createConnectQueryKey(getAccountOnboardingConfig, {
              accountId: account.id,
            }),
            new GetAccountOnboardingConfigResponse({
              config: resp.config,
            })
          );
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
        queryclient.setQueryData(
          createConnectQueryKey(getConnection, {
            id: connection.connection.id,
          }),
          new GetConnectionResponse({
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
              config.connectionOptions?.maxConnectionLimit ?? 80,
          },
          tunnel: {
            host: config.tunnel?.host ?? '',
            port: config.tunnel?.port ?? 22,
            knownHostPublicKey: config.tunnel?.knownHostPublicKey ?? '',
            user: config.tunnel?.user ?? '',
            passphrase: passPhrase,
            privateKey: privateKey,
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
        toast({
          title: 'Unable to retrieve connection data for clone!',
          description: getErrorMessage(error),
          variant: 'destructive',
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
                    placeholder="mysql://username:password@hostname:port/database"
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
                  value={field.value ? field.value.toString() : 80}
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
                setIsValidating(false);
                setValidationResponse(res);
              } catch (err) {
                setValidationResponse(
                  new CheckConnectionConfigResponse({
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
        {validationResponse && validationResponse.isConnected && (
          <SuccessAlert description={'Successfully connected!'} />
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
