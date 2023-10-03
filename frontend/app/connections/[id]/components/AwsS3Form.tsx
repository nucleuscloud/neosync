'use client';
import RequiredLabel from '@/components/labels/RequiredLabel';
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
    roleArn: Yup.string().optional(),
    accessKeyId: Yup.string(),
    accessKey: Yup.string()
      .ensure()
      .when('accessKeyId', {
        is: (val?: string) => {
          return val && val != '';
        },
        then: (schema) => schema.required(),
      }),
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
    defaultValues,
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
          name="s3.roleArn"
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
          name="s3.accessKeyId"
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
          name="s3.accessKey"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input placeholder="Access Key" {...field} />
              </FormControl>
              <FormDescription>Access Key</FormDescription>
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
  console.log(s3);
  const credentials =
    s3.accessKeyId && s3.accessKey
      ? new AwsS3Credentials({
          accessKeyId: s3.accessKeyId,
          accessKey: s3.accessKey,
        })
      : undefined;
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
              roleArn: s3.roleArn,
              credentials,
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
