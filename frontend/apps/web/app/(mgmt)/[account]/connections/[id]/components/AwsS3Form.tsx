'use client';
import ButtonText from '@/components/ButtonText';
import { PasswordInput } from '@/components/PasswordComponent';
import Spinner from '@/components/Spinner';
import RequiredLabel from '@/components/labels/RequiredLabel';
import { useAccount } from '@/components/providers/account-provider';
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
import {
  AWSFormValues,
  AWS_FORM_SCHEMA,
  EditConnectionFormContext,
} from '@/yup-validations/connections';
import { create } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  ConnectionService,
  UpdateConnectionRequestSchema,
  UpdateConnectionResponse,
} from '@neosync/sdk';
import { useForm } from 'react-hook-form';
import { IoAlertCircleOutline } from 'react-icons/io5';
import { buildConnectionConfigAwsS3 } from '../../util';

interface Props {
  connectionId: string;
  defaultValues: AWSFormValues;
  onSaved(updatedConnectionResp: UpdateConnectionResponse): void;
  onSaveFailed(err: unknown): void;
}

export default function AwsS3Form(props: Props) {
  const { connectionId, defaultValues, onSaved, onSaveFailed } = props;
  const { account } = useAccount();
  const { mutateAsync: isConnectionNameAvailableAsync } = useMutation(
    ConnectionService.method.isConnectionNameAvailable
  );
  const form = useForm<AWSFormValues, EditConnectionFormContext>({
    resolver: yupResolver(AWS_FORM_SCHEMA),
    defaultValues: {
      connectionName: '',
      s3: {},
    },
    values: defaultValues,
    context: {
      originalConnectionName: defaultValues.connectionName,
      accountId: account?.id ?? '',
      isConnectionNameAvailable: isConnectionNameAvailableAsync,
    },
  });
  const { mutateAsync } = useMutation(
    ConnectionService.method.updateConnection
  );

  async function onSubmit(values: AWSFormValues) {
    try {
      const connectionResp = await mutateAsync(
        create(UpdateConnectionRequestSchema, {
          id: connectionId,
          name: values.connectionName,
          connectionConfig: buildConnectionConfigAwsS3(values),
        })
      );
      onSaved(connectionResp);
    } catch (err) {
      console.error(err);
      onSaveFailed(err);
    }
  }
  return (
    <div className="flex flex-col gap-4">
      <Alert variant="warning">
        <div className="flex flex-row items-center gap-2">
          <IoAlertCircleOutline className="h-6 w-6" />
          <AlertTitle className="font-semibold">Heads up!</AlertTitle>
        </div>
        <AlertDescription className="pl-8">
          Right now AWS S3 connections can only be used as a destination
        </AlertDescription>
      </Alert>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-2">
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
                  <PasswordInput placeholder="Secret Access Key" {...field} />
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
                leftIcon={
                  form.formState.isSubmitting ? <Spinner /> : <div></div>
                }
                text="Update"
              />
            </Button>
          </div>
        </form>
      </Form>
    </div>
  );
}
