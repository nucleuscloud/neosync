'use client';
import ButtonText from '@/components/ButtonText';
import FormError from '@/components/FormError';
import Spinner from '@/components/Spinner';
import RequiredLabel from '@/components/labels/RequiredLabel';
import { setOnboardingConfig } from '@/components/onboarding-checklist/OnboardingChecklist';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import SwitchCard from '@/components/switches/SwitchCard';
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
import { toast } from '@/components/ui/use-toast';
import { useGetAccountOnboardingConfig } from '@/libs/hooks/useGetAccountOnboardingConfig';
import { AWSFormValues, AWS_FORM_SCHEMA } from '@/yup-validations/connections';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  AwsS3ConnectionConfig,
  AwsS3Credentials,
  ConnectionConfig,
  CreateConnectionRequest,
  CreateConnectionResponse,
  GetAccountOnboardingConfigResponse,
} from '@neosync/sdk';
import { useRouter, useSearchParams } from 'next/navigation';
import { useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { IoAlertCircleOutline } from 'react-icons/io5';
import { GetConnectionCloneValues } from '../postgres/PostgresForm';

export default function AwsS3Form() {
  const searchParams = useSearchParams();
  const { account } = useAccount();
  const sourceConnId = searchParams.get('sourceId');
  const [isLoading, setIsLoading] = useState<boolean>();
  const { data: onboardingData, mutate } = useGetAccountOnboardingConfig(
    account?.id ?? ''
  );
  const form = useForm<AWSFormValues>({
    resolver: yupResolver(AWS_FORM_SCHEMA),
    defaultValues: {
      connectionName: '',
      s3: {
        bucket: '',
      },
    },
    context: { accountId: account?.id ?? '' },
  });
  const router = useRouter();

  async function onSubmit(values: AWSFormValues) {
    if (!account) {
      return;
    }
    try {
      const connection = await createAwsS3Connection(
        values.s3,
        values.connectionName,
        account.id
      );

      // updates the onboarding data
      if (
        onboardingData?.config?.hasCreatedSourceConnection &&
        !onboardingData?.config.hasCreatedDestinationConnection
      ) {
        try {
          const resp = await setOnboardingConfig(account.id, {
            hasCreatedSourceConnection:
              onboardingData.config.hasCreatedSourceConnection,
            hasCreatedDestinationConnection: true,
            hasCreatedJob: onboardingData.config.hasCreatedJob,
            hasInvitedMembers: onboardingData.config.hasInvitedMembers,
          });
          mutate(
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
          const resp = await setOnboardingConfig(account.id, {
            hasCreatedSourceConnection: true,
            hasCreatedDestinationConnection:
              onboardingData?.config?.hasCreatedSourceConnection ?? true,
            hasCreatedJob: onboardingData?.config?.hasCreatedJob ?? true,
            hasInvitedMembers:
              onboardingData?.config?.hasInvitedMembers ?? true,
          });
          mutate(
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
        router.push(
          `/${account?.name}/connections/${connection.connection.id}`
        );
      } else {
        router.push(`/${account?.name}/connections`);
      }
    } catch (err) {
      console.error(err);
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
            connData.connection?.connectionConfig?.config.case === 'awsS3Config'
          ) {
            const s3Config =
              connData.connection?.connectionConfig?.config.value;

            /* reset the form with the new values and include the fallback values because of our validation schema requires a string and not undefined which is okay because it will tell the user that something is wrong instead of the user not realizing that it's undefined
             */
            form.reset({
              ...form.getValues(),
              connectionName: connData.connection?.name + '-copy',
              s3: {
                bucket: s3Config.bucket ?? '',
                pathPrefix: s3Config.pathPrefix ?? '',
                region: s3Config.region ?? '',
                endpoint: s3Config.endpoint ?? '',
                credentials: {
                  profile: s3Config.credentials?.profile ?? '',
                  accessKeyId: s3Config.credentials?.accessKeyId ?? '',
                  secretAccessKey: s3Config.credentials?.secretAccessKey ?? '',
                  sessionToken: s3Config.credentials?.sessionToken ?? '',
                  fromEc2Role: s3Config.credentials?.fromEc2Role ?? false,
                  roleArn: s3Config.credentials?.roleArn ?? '',
                  roleExternalId: s3Config.credentials?.roleExternalId ?? '',
                },
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
    return <SkeletonForm />;
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <Alert variant="warning">
          <div className="flex flex-row items-center gap-2">
            <IoAlertCircleOutline className="h-6 w-6" />
            <AlertTitle className="font-semibold">Heads up!</AlertTitle>
          </div>
          <AlertDescription className="pl-8">
            Right now AWS S3 connections can only be used as a destination
          </AlertDescription>
        </Alert>
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
          name="s3.bucket"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                <RequiredLabel />
                Bucket
              </FormLabel>
              <FormDescription>The bucket</FormDescription>
              <FormControl>
                <Input placeholder="Bucket" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="s3.pathPrefix"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Path Prefix</FormLabel>
              <FormDescription>The path prefix of the bucket</FormDescription>
              <FormControl>
                <Input placeholder="/..." {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="s3.region"
          render={({ field }) => (
            <FormItem>
              <FormLabel>AWS Region</FormLabel>
              <FormDescription>The AWS region to target</FormDescription>
              <FormControl>
                <Input placeholder="" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="s3.endpoint"
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
        <FormField
          control={form.control}
          name="s3.credentials.profile"
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

        <FormField
          control={form.control}
          name="s3.credentials.accessKeyId"
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
          name="s3.credentials.secretAccessKey"
          render={({ field }) => (
            <FormItem>
              <FormLabel>AWS Secret Access Key</FormLabel>
              <FormDescription>AWS Secret Access Key</FormDescription>
              <FormControl>
                <Input placeholder="Secret Access Key" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="s3.credentials.sessionToken"
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

        <FormField
          control={form.control}
          name="s3.credentials.fromEc2Role"
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

        <FormField
          control={form.control}
          name="s3.credentials.roleArn"
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
          name="s3.credentials.roleExternalId"
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
        <div className="flex flex-row gap-3 justify-end">
          <Button type="submit">
            <ButtonText
              leftIcon={form.formState.isSubmitting ? <Spinner /> : <div></div>}
              text="Submit"
            />
          </Button>
        </div>
      </form>
    </Form>
  );
}

async function createAwsS3Connection(
  s3: AWSFormValues['s3'],
  name: string,
  accountId: string
): Promise<CreateConnectionResponse> {
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
            case: 'awsS3Config',
            value: new AwsS3ConnectionConfig({
              bucket: s3.bucket,
              pathPrefix: s3.pathPrefix,
              region: s3.region,
              endpoint: s3.endpoint,
              credentials: new AwsS3Credentials({
                profile: s3.credentials?.profile,
                accessKeyId: s3.credentials?.accessKeyId,
                secretAccessKey: s3.credentials?.secretAccessKey,
                fromEc2Role: s3.credentials?.fromEc2Role,
                roleArn: s3.credentials?.roleArn,
                roleExternalId: s3.credentials?.roleExternalId,
                sessionToken: s3.credentials?.sessionToken,
              }),
            }),
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
