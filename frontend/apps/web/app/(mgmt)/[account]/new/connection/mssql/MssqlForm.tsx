'use client';
import ButtonText from '@/components/ButtonText';
import FormError from '@/components/FormError';
import Spinner from '@/components/Spinner';
import RequiredLabel from '@/components/labels/RequiredLabel';
import { buildAccountOnboardingConfig } from '@/components/onboarding-checklist/OnboardingChecklist';
import PermissionsDialog from '@/components/permissions/PermissionsDialog';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
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
import { getErrorMessage } from '@/util/util';
import {
  MssqlCreateConnectionFormContext,
  MssqlFormValues,
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
  isConnectionNameAvailable,
  setAccountOnboardingConfig,
} from '@neosync/sdk/connectquery';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useQueryClient } from '@tanstack/react-query';
import { useRouter, useSearchParams } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
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
        maxConnectionLimit: 80,
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
  const { data: onboardingData } = useQuery(
    getAccountOnboardingConfig,
    { accountId: account?.id ?? '' },
    { enabled: !!account?.id }
  );
  const queryclient = useQueryClient();
  const { mutateAsync: setOnboardingConfigAsync } = useMutation(
    setAccountOnboardingConfig
  );

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
          toast.error('Unable to update onboarding status!', {
            description: getErrorMessage(e),
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
          toast.error('Unable to update onboarding status!', {
            description: getErrorMessage(e),
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

        const dbConfig = {
          url: '',
        };

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
              config.connectionOptions?.maxConnectionLimit ?? 80,
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
