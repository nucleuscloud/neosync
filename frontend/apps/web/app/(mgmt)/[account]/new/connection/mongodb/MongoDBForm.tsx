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
  CreateConnectionFormContext,
  MongoDbFormValues,
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
import { buildConnectionConfigMongo } from '../../../connections/util';

export default function MongoDBForm(): ReactElement {
  const searchParams = useSearchParams();
  const { account } = useAccount();
  const sourceConnId = searchParams.get('sourceId');
  const [isLoading, setIsLoading] = useState<boolean>();
  const queryclient = useQueryClient();
  const { mutateAsync: isConnectionNameAvailableAsync } = useMutation(
    isConnectionNameAvailable
  );
  const form = useForm<MongoDbFormValues, CreateConnectionFormContext>({
    resolver: yupResolver(MongoDbFormValues),
    mode: 'onChange',
    defaultValues: {
      connectionName: '',
      url: '',

      clientTls: {
        rootCert: '',
        clientCert: '',
        clientKey: '',
      },
    },
    context: {
      accountId: account?.id ?? '',
      isConnectionNameAvailable: isConnectionNameAvailableAsync,
    },
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
  const { mutateAsync: createMongoDbConnection } =
    useMutation(createConnection);
  const { mutateAsync: checkMongoDbConnection } = useMutation(
    checkConnectionConfig
  );
  const { mutateAsync: getMongoDbConnection } = useMutation(getConnection);

  useEffect(() => {
    const fetchData = async () => {
      if (!sourceConnId || !account?.id) {
        return;
      }
      setIsLoading(true);
      try {
        const connData = await getMongoDbConnection({ id: sourceConnId });
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
        toast.error('Unable to retrieve connection data for clone!', {
          description: getErrorMessage(error),
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
      const newConnection = await createMongoDbConnection({
        name: values.connectionName,
        accountId: account.id,
        connectionConfig: buildConnectionConfigMongo(values),
      });
      posthog.capture('New Connection Created', { type: 'mongodb' });
      toast.success('Successfully created connection!');
      const returnTo = searchParams.get('returnTo');
      if (returnTo) {
        router.push(returnTo);
      } else if (newConnection.connection?.id) {
        queryclient.setQueryData(
          createConnectQueryKey(getConnection, {
            id: newConnection.connection.id,
          }),
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
      toast.error('Unable to create connection!', {
        description: getErrorMessage(err),
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
      const res = await checkMongoDbConnection({
        connectionConfig: buildConnectionConfigMongo(form.getValues()),
      });
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
            <AccordionTrigger>Client TLS Certificates</AccordionTrigger>
            <AccordionContent className="flex flex-col gap-4 p-2">
              <div className="text-sm">
                Configuring this section allows Neosync to connect to the
                database using SSL/TLS.
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
