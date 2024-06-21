'use client';
import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import RequiredLabel from '@/components/labels/RequiredLabel';
import { setOnboardingConfig } from '@/components/onboarding-checklist/OnboardingChecklist';
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
import { useToast } from '@/components/ui/use-toast';
import { useGetAccountOnboardingConfig } from '@/libs/hooks/useGetAccountOnboardingConfig';
import { getConnection } from '@/libs/hooks/useGetConnection';
import { getErrorMessage } from '@/util/util';
import {
  MongoDbFormContext,
  MongoDbFormValues,
} from '@/yup-validations/connections';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckConnectionConfigResponse,
  GetAccountOnboardingConfigResponse,
  GetConnectionResponse,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useRouter, useSearchParams } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { mutate } from 'swr';
import {
  checkMongoConnection,
  createMongoConnection,
} from '../../../connections/util';

export default function MongoDBForm(): ReactElement {
  const searchParams = useSearchParams();
  const { account } = useAccount();
  const sourceConnId = searchParams.get('sourceId');
  const [isLoading, setIsLoading] = useState<boolean>();
  const { data: onboardingData, mutate: mutateOnboardingData } =
    useGetAccountOnboardingConfig(account?.id ?? '');

  const form = useForm<MongoDbFormValues, MongoDbFormContext>({
    resolver: yupResolver(MongoDbFormValues),
    mode: 'onChange',
    defaultValues: {
      connectionName: '',
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
