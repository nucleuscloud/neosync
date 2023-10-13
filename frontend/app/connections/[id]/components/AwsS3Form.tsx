'use client';
import RequiredLabel from '@/components/labels/RequiredLabel';
import SwitchCard from '@/components/switches/SwitchCard';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import {
  AwsS3ConnectionConfig,
  AwsS3Credentials,
  ConnectionConfig,
  UpdateConnectionRequest,
  UpdateConnectionResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { yupResolver } from '@hookform/resolvers/yup';
import { useForm } from 'react-hook-form';
import { FaTerminal } from 'react-icons/fa';
import * as Yup from 'yup';

const FORM_SCHEMA = Yup.object({
  connectionName: Yup.string(),
  s3: Yup.object({
    bucketArn: Yup.string().required(),
    pathPrefix: Yup.string().optional(),
    region: Yup.string().optional(),
    endpoint: Yup.string().optional(),
    credentials: Yup.object({
      profile: Yup.string().optional(),
      accessKeyId: Yup.string(),
      secretAccessKey: Yup.string().optional(),
      sessionToken: Yup.string().optional(),
      fromEc2Role: Yup.boolean().optional(),
      roleArn: Yup.string().optional(),
      roleExternalId: Yup.string().optional(),
    }).optional(),
  }).required(),
});

type FormValues = Yup.InferType<typeof FORM_SCHEMA>;

interface Props {
  connectionId: string;
  defaultValues: FormValues;
  onSaved(updatedConnectionResp: UpdateConnectionResponse): void;
  onSaveFailed(err: unknown): void;
}

export default function AwsS3Form(props: Props) {
  const { connectionId, defaultValues, onSaved, onSaveFailed } = props;
  const form = useForm<FormValues>({
    resolver: yupResolver(FORM_SCHEMA),
    defaultValues: {
      connectionName: '',
      s3: {},
    },
    values: defaultValues,
  });

  async function onSubmit(values: FormValues) {
    try {
      const connectionResp = await updateAwsS3Connection(
        values.s3,
        connectionId
      );
      onSaved(connectionResp);
    } catch (err) {
      console.error(err);
      onSaveFailed(err);
    }
  }
  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <Alert>
          <FaTerminal className="h-4 w-4" />
          <AlertTitle>Heads up!</AlertTitle>
          <AlertDescription>
            Right now AWS S3 connections can only be used as a destination
          </AlertDescription>
        </Alert>
        <FormField
          control={form.control}
          name="connectionName"
          disabled={true}
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input placeholder="Connection Name" {...field} />
              </FormControl>
              <FormDescription>
                <RequiredLabel />
                The unique name of the connection.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="s3.bucketArn"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input placeholder="Bucket ARN" {...field} />
              </FormControl>
              <FormDescription>
                <RequiredLabel />
                Bucket ARN
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="s3.pathPrefix"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input placeholder="/..." {...field} />
              </FormControl>
              <FormDescription>Path Prefix</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="s3.region"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input placeholder="" {...field} />
              </FormControl>
              <FormDescription>The AWS region to target</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="s3.endpoint"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input placeholder="" {...field} />
              </FormControl>
              <FormDescription>
                Allows specifying a custom endpoint for the AWS API
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="space-y-2">
          <h2 className="text-1xl font-bold tracking-tight">Manual Setup</h2>
          <p className="text-sm tracking-tight">
            Optional manual configuration of AWS credentials to use
          </p>
        </div>

        <FormField
          control={form.control}
          name="s3.credentials.profile"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input placeholder="default" {...field} />
              </FormControl>
              <FormDescription>AWS Profile Name</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="s3.credentials.accessKeyId"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input placeholder="Access Key Id" {...field} />
              </FormControl>
              <FormDescription>Access Key Id</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="s3.credentials.secretAccessKey"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input placeholder="Secret Access Key" {...field} />
              </FormControl>
              <FormDescription>Secret Access Key</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="s3.credentials.sessionToken"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input placeholder="Session Token" {...field} />
              </FormControl>
              <FormDescription>Session Token</FormDescription>
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
              {/* <FormDescription>From EC2 Role</FormDescription> */}
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="s3.credentials.roleArn"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input placeholder="Role Arn" {...field} />
              </FormControl>
              <FormDescription>Role Arn</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="s3.credentials.roleExternalId"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input placeholder="Role External Id" {...field} />
              </FormControl>
              <FormDescription>Role External Id</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className="flex flex-row gap-3 justify-items-end">
          <Button type="submit">Submit</Button>
        </div>
      </form>
    </Form>
  );
}

async function updateAwsS3Connection(
  s3: FormValues['s3'],
  connectionId: string
): Promise<UpdateConnectionResponse> {
  const res = await fetch(`/api/connections/${connectionId}`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new UpdateConnectionRequest({
        id: connectionId,
        connectionConfig: new ConnectionConfig({
          config: {
            case: 'awsS3Config',
            value: new AwsS3ConnectionConfig({
              bucketArn: s3.bucketArn,
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
  return UpdateConnectionResponse.fromJson(await res.json());
}
