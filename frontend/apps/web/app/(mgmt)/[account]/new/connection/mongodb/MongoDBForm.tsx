'use client';
import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import RequiredLabel from '@/components/labels/RequiredLabel';
import { setOnboardingConfig } from '@/components/onboarding-checklist/OnboardingChecklist';
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
import { useToast } from '@/components/ui/use-toast';
import { useGetAccountOnboardingConfig } from '@/libs/hooks/useGetAccountOnboardingConfig';
import { getConnection } from '@/libs/hooks/useGetConnection';
import { getErrorMessage } from '@/util/util';
import { MongoDbFormValues } from '@/yup-validations/connections';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckConnectionConfigResponse,
  ConnectionConfig,
  CreateConnectionRequest,
  CreateConnectionResponse,
  GetAccountOnboardingConfigResponse,
  GetConnectionResponse,
  MongoConnectionConfig,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useRouter, useSearchParams } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { mutate } from 'swr';
import { checkMongoConnection } from '../../../connections/util';

export default function MongoDBForm(): ReactElement {
  const searchParams = useSearchParams();
  const { account } = useAccount();
  const sourceConnId = searchParams.get('sourceId');
  const [isLoading, setIsLoading] = useState<boolean>();
  const { data: onboardingData, mutate: mutateOnboardingData } =
    useGetAccountOnboardingConfig(account?.id ?? '');

  const form = useForm<MongoDbFormValues>({
    resolver: yupResolver(MongoDbFormValues),
    mode: 'onChange',
    defaultValues: {
      connectionName: '',
      url: '',
    },
    context: { accountId: account?.id ?? '' },
  });

  const router = useRouter();
  const [isValidating, setIsValidating] = useState<boolean>(false);
  const [validationResponse, setValidationResponse] = useState<
    CheckConnectionConfigResponse | undefined
  >();
  const [openPermissionDialog, setOpenPermissionDialog] =
    useState<boolean>(false);
  const [isSubmitting, setIsSubmitting] = useState<boolean>();
  const posthog = usePostHog();
  const { toast } = useToast();

  useEffect(() => {
    const fetchData = async () => {
      if (!sourceConnId || !account?.id) {
        return;
      }
      setIsLoading(true);
      try {
        const connData = await getConnection(account.id, sourceConnId);
        if (
          connData.connection?.connectionConfig?.config.case !== 'mongoConfig'
        ) {
          return;
        }

        const config = connData.connection?.connectionConfig?.config.value;
        const mongoConnConfigValue = config.connectionConfig.value;

        form.reset({
          ...form.getValues(),
          connectionName: connData.connection?.name + '-copy',
          url: mongoConnConfigValue ?? '',
        });
      } catch (error) {
        console.error('Failed to fetch connection data:', error);
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

  async function onSubmit(values: MongoDbFormValues): Promise<void> {
    if (!account || isSubmitting) {
      return;
    }
    setIsSubmitting(true);
    try {
      const newConnection = await createMongoConnection(values, account.id);
      posthog.capture('New Connection Created', { type: 'mongodb' });
      toast({
        title: 'Successfully created connection!',
        variant: 'success',
      });

      // updates the onboarding data
      try {
        const resp = await setOnboardingConfig(account.id, {
          hasCreatedSourceConnection:
            onboardingData?.config?.hasCreatedSourceConnection ?? true,
          hasCreatedDestinationConnection:
            onboardingData?.config?.hasCreatedDestinationConnection ??
            onboardingData?.config?.hasCreatedSourceConnection ??
            false,
          hasCreatedJob: onboardingData?.config?.hasCreatedJob ?? false,
          hasInvitedMembers: onboardingData?.config?.hasInvitedMembers ?? false,
        });
        mutateOnboardingData(
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

      const returnTo = searchParams.get('returnTo');
      if (returnTo) {
        router.push(returnTo);
      } else if (newConnection.connection?.id) {
        mutate(
          `$/{account?.name}/connections/${newConnection.connection.id}`,
          new GetConnectionResponse({
            connection: newConnection.connection,
          })
        );
        router.push(
          `/${account?.name}/connections/${newConnection.connection.id}`
        );
      } else {
        router.push(`/${account.name}/connections`);
      }
    } catch (err) {
      toast({
        title: 'Unable to create connection',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  async function onValidationClick(): Promise<void> {
    if (isValidating) {
      return;
    }
    setIsValidating(true);
    try {
      const res = await checkMongoConnection(
        form.getValues(),
        account?.id ?? ''
      );
      setValidationResponse(res);
      setOpenPermissionDialog(!!res.isConnected);
    } catch (err) {
      setValidationResponse(
        new CheckConnectionConfigResponse({
          isConnected: false,
          connectionError: err instanceof Error ? err.message : 'unknown error',
        })
      );
    } finally {
      setIsValidating(false);
    }
  }

  if (isLoading || !account?.id) {
    return <SkeletonForm />;
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <FormField
          control={form.control}
          name="connectionName"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                Connection Name
              </FormLabel>
              <FormDescription>
                The unique name of the connection
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
          name="url"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                Connection URL
              </FormLabel>
              <FormDescription>The url fo the MongoDB server</FormDescription>
              <FormControl>
                <Input {...field} />
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
          connectionType="mongodb"
        />

        <div className="flex flex-row gap-3 justify-between">
          <Button
            type="button"
            variant="outline"
            onClick={() => onValidationClick()}
          >
            <ButtonText
              leftIcon={
                isValidating ? (
                  <Spinner className="text-black dark:text-white" />
                ) : (
                  <div />
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

async function createMongoConnection(
  values: MongoDbFormValues,
  accountId: string
): Promise<CreateConnectionResponse> {
  const mongoconfig = new MongoConnectionConfig({
    connectionConfig: {
      case: 'url',
      value: values.url,
    },
    clientTls: undefined,
    tunnel: undefined,
  });

  const res = await fetch(`/api/accounts/${accountId}/connections`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new CreateConnectionRequest({
        accountId,
        name: values.connectionName,
        connectionConfig: new ConnectionConfig({
          config: {
            case: 'mongoConfig',
            value: mongoconfig,
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
