'use client';
import ButtonText from '@/components/ButtonText';
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
import { Textarea } from '@/components/ui/textarea';
import { getErrorMessage } from '@/util/util';
import {
  MssqlCreateConnectionFormContext,
  MssqlFormValues,
} from '@/yup-validations/connections';
import { createConnectQueryKey, useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckConnectionConfigResponse,
  GetConnectionResponse,
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
import { buildConnectionConfigMssql } from '../../../connections/util';

export default function MssqlForm() {
  const { account } = useAccount();
  const searchParams = useSearchParams();
  const sourceConnId = searchParams.get('sourceId');
  const [isLoading, setIsLoading] = useState<boolean>();

  const { mutateAsync: isConnectionNameAvailableAsync } = useMutation(
    isConnectionNameAvailable
  );
  const form = useForm<MssqlFormValues, MssqlCreateConnectionFormContext>({
    resolver: yupResolver(MssqlFormValues),
    defaultValues: {
      connectionName: '',
      db: {
        url: '',
      },
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
    },
    context: {
      accountId: account?.id ?? '',
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
  const { mutateAsync: createMssqlConnection } = useMutation(createConnection);
  const { mutateAsync: checkMssqlConnection } = useMutation(
    checkConnectionConfig
  );
  const { mutateAsync: getMssqlConnection } = useMutation(getConnection);
  const queryclient = useQueryClient();
  async function onSubmit(values: MssqlFormValues) {
    if (!account) {
      return;
    }
    try {
      const connection = await createMssqlConnection({
        name: values.connectionName,
        accountId: account.id,
        connectionConfig: buildConnectionConfigMssql(values),
      });
      posthog.capture('New Connection Created', { type: 'mssql' });
      toast.success('Successfully created connection!');
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
        const connData = await getMssqlConnection({ id: sourceConnId });
        if (
          connData.connection?.connectionConfig?.config.case !== 'mssqlConfig'
        ) {
          return;
        }

        const config = connData.connection?.connectionConfig?.config.value;
        const mssqlConfig = config.connectionConfig.value;

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
          db: {
            url: typeof mssqlConfig === 'string' ? mssqlConfig : '',
          },
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
        });
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

        <FormField
          control={form.control}
          name="db.url"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                Connection URL
              </FormLabel>
              <FormDescription>
                Your connection URL in URL format
              </FormDescription>
              <FormControl>
                <Input
                  placeholder="sqlserver://username:password@host:port/instance?param1=value&param2=value"
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="flex flex-col gap-0">
          <FormField
            control={form.control}
            name="options.maxConnectionLimit"
            render={({ field }) => (
              <FormItem>
                <div className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                  <div className="space-y-0.5">
                    <FormLabel>Max Connection Limit</FormLabel>
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
            validationResponse ?? new CheckConnectionConfigResponse({})
          }
          openPermissionDialog={openPermissionDialog}
          setOpenPermissionDialog={setOpenPermissionDialog}
          isValidating={isValidating}
          connectionName={form.getValues('connectionName')}
          connectionType="mssql"
        />
        <div className="flex flex-row gap-3 justify-between">
          <Button
            variant="outline"
            disabled={!form.formState.isValid}
            onClick={async () => {
              setIsValidating(true);
              const values = form.getValues();
              try {
                const res = await checkMssqlConnection({
                  connectionConfig: buildConnectionConfigMssql(values),
                });
                setValidationResponse(res);
                setOpenPermissionDialog(!!res?.isConnected);
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
