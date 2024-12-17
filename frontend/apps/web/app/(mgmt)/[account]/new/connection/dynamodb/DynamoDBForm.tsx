'use client';
import ButtonText from '@/components/ButtonText';
import { PasswordInput } from '@/components/PasswordComponent';
import Spinner from '@/components/Spinner';
import RequiredLabel from '@/components/labels/RequiredLabel';
import PermissionsDialog from '@/components/permissions/PermissionsDialog';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import SwitchCard from '@/components/switches/SwitchCard';
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
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { getErrorMessage } from '@/util/util';
import {
  CreateConnectionFormContext,
  DynamoDbFormValues,
} from '@/yup-validations/connections';
import { create } from '@bufbuild/protobuf';
import { createConnectQueryKey, useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CheckConnectionConfigResponse,
  CheckConnectionConfigResponseSchema,
  ConnectionService,
  GetConnectionResponseSchema,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useQueryClient } from '@tanstack/react-query';
import { useRouter, useSearchParams } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { buildConnectionConfigDynamoDB } from '../../../connections/util';

export default function NewDynamoDBForm(): ReactElement {
  const searchParams = useSearchParams();
  const { account } = useAccount();
  const { data: systemAppConfig } = useGetSystemAppConfig();
  const sourceConnId = searchParams.get('sourceId');
  const [isLoading, setIsLoading] = useState<boolean>();
  const queryclient = useQueryClient();
  const { mutateAsync: isConnectionNameAvailableAsync } = useMutation(
    ConnectionService.method.isConnectionNameAvailable
  );
  const form = useForm<DynamoDbFormValues, CreateConnectionFormContext>({
    resolver: yupResolver(DynamoDbFormValues),
    mode: 'onChange',
    defaultValues: {
      connectionName: '',

      db: {
        endpoint: '',
        region: '',
        credentials: {
          accessKeyId: '',
          fromEc2Role: false,
          profile: '',
          roleArn: '',
          roleExternalId: '',
          secretAccessKey: '',
          sessionToken: '',
        },
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
  const { mutateAsync: createDbConnection } = useMutation(
    ConnectionService.method.createConnection
  );
  const { mutateAsync: checkDbConnection } = useMutation(
    ConnectionService.method.checkConnectionConfig
  );
  const { mutateAsync: getDbConnection } = useMutation(
    ConnectionService.method.getConnection
  );

  useEffect(() => {
    const fetchData = async () => {
      if (!sourceConnId || !account?.id) {
        return;
      }
      setIsLoading(true);
      try {
        const connData = await getDbConnection({ id: sourceConnId });
        if (
          connData.connection?.connectionConfig?.config.case !==
          'dynamodbConfig'
        ) {
          return;
        }

        const config = connData.connection?.connectionConfig?.config.value;

        form.reset({
          ...form.getValues(),
          connectionName: connData.connection?.name + '-copy',
          db: {
            region: config.region ?? '',
            endpoint: config.endpoint ?? '',
            credentials: {
              profile: config.credentials?.profile ?? '',
              accessKeyId: config.credentials?.accessKeyId ?? '',
              secretAccessKey: config.credentials?.secretAccessKey ?? '',
              sessionToken: config.credentials?.sessionToken ?? '',
              fromEc2Role: config.credentials?.fromEc2Role ?? false,
              roleArn: config.credentials?.roleArn ?? '',
              roleExternalId: config.credentials?.roleExternalId ?? '',
            },
          },
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

  async function onSubmit(values: DynamoDbFormValues): Promise<void> {
    if (!account || isSubmitting) {
      return;
    }
    setIsSubmitting(true);
    try {
      const newConnection = await createDbConnection({
        name: values.connectionName,
        accountId: account.id,
        connectionConfig: buildConnectionConfigDynamoDB(values),
      });
      posthog.capture('New Connection Created', { type: 'dynamodb' });
      toast.success('Successfully created connection!');

      const returnTo = searchParams.get('returnTo');
      if (returnTo) {
        router.push(returnTo);
      } else if (newConnection.connection?.id) {
        queryclient.setQueryData(
          createConnectQueryKey({
            schema: ConnectionService.method.getConnection,
            input: { id: newConnection.connection.id },
            cardinality: undefined,
          }),
          create(GetConnectionResponseSchema, {
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
      toast.error('Unable to create connection', {
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
      const res = await checkDbConnection({
        connectionConfig: buildConnectionConfigDynamoDB(form.getValues()),
      });
      setValidationResponse(res);
      setOpenPermissionDialog(!!res.isConnected);
    } catch (err) {
      setValidationResponse(
        create(CheckConnectionConfigResponseSchema, {
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

        <Accordion
          type="multiple"
          defaultValue={['credentials']}
          className="w-full"
        >
          <AccordionItem value="credentials">
            <AccordionTrigger>AWS Credentials</AccordionTrigger>
            <AccordionContent className="space-y-4 p-2">
              <p className="text-sm">
                This section is used to configure authentication credentials to
                allow access to the DynamoDB tables and will be specific to how
                you wish Neosync to connect to Dynamo.
              </p>
              <div className="space-y-8">
                {!systemAppConfig?.isNeosyncCloud && (
                  <FormField
                    control={form.control}
                    name="db.credentials.profile"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>AWS Profile Name</FormLabel>
                        <FormDescription>AWS Profile Name</FormDescription>
                        <FormControl>
                          <Input placeholder="default" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}

                <FormField
                  control={form.control}
                  name="db.credentials.accessKeyId"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Access Key Id</FormLabel>
                      <FormDescription>Access Key Id</FormDescription>
                      <FormControl>
                        <Input placeholder="Access Key Id" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="db.credentials.secretAccessKey"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>AWS Secret Access Key</FormLabel>
                      <FormDescription>AWS Secret Access Key</FormDescription>
                      <FormControl>
                        <PasswordInput
                          placeholder="Secret Access Key"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="db.credentials.sessionToken"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>AWS Session Token</FormLabel>
                      <FormDescription>AWS Session Token</FormDescription>
                      <FormControl>
                        <Input placeholder="Session Token" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                {!systemAppConfig?.isNeosyncCloud && (
                  <FormField
                    control={form.control}
                    name="db.credentials.fromEc2Role"
                    render={({ field }) => (
                      <FormItem>
                        <FormControl>
                          <SwitchCard
                            isChecked={field.value || false}
                            onCheckedChange={field.onChange}
                            title="From EC2 Role"
                            description="Use the credentials of a host EC2 machine configured to assume an IAM role associated with the instance."
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}

                <FormField
                  control={form.control}
                  name="db.credentials.roleArn"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>AWS Role ARN</FormLabel>
                      <FormDescription>Role ARN</FormDescription>
                      <FormControl>
                        <Input placeholder="Role Arn" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="db.credentials.roleExternalId"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>AWS Role External Id</FormLabel>
                      <FormDescription>Role External Id</FormDescription>
                      <FormControl>
                        <Input placeholder="Role External Id" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </AccordionContent>
          </AccordionItem>

          <AccordionItem value="advanced">
            <AccordionTrigger>AWS Advanced Configuration</AccordionTrigger>
            <AccordionContent className="space-y-4 p-2">
              <p className="text-sm">
                This is an optional section and is used if you need to tweak the
                AWS SDK to connect to a different region or endpoint other than
                the default.
              </p>
              <div className="space-y-8">
                <FormField
                  control={form.control}
                  name="db.region"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>AWS Region</FormLabel>
                      <FormDescription>
                        The AWS region to target
                      </FormDescription>
                      <FormControl>
                        <Input placeholder="" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="db.endpoint"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Custom Endpoint</FormLabel>
                      <FormDescription>
                        Allows specifying a custom endpoint for the AWS API
                      </FormDescription>
                      <FormControl>
                        <Input placeholder="" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
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
          connectionType="dynamodb"
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
