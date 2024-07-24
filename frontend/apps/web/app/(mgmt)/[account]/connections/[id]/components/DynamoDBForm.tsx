'use client';
import ButtonText from '@/components/ButtonText';
import FormError from '@/components/FormError';
import { PasswordInput } from '@/components/PasswordComponent';
import Spinner from '@/components/Spinner';
import RequiredLabel from '@/components/labels/RequiredLabel';
import { useAccount } from '@/components/providers/account-provider';
import SwitchCard from '@/components/switches/SwitchCard';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
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
  DynamoDbFormValues,
  EditConnectionFormContext,
} from '@/yup-validations/connections';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  UpdateConnectionRequest,
  UpdateConnectionResponse,
} from '@neosync/sdk';
import {
  isConnectionNameAvailable,
  updateConnection,
} from '@neosync/sdk/connectquery';
import { Controller, useForm } from 'react-hook-form';
import { buildConnectionConfigDynamoDB } from '../../util';

interface Props {
  connectionId: string;
  defaultValues: DynamoDbFormValues;
  onSaved(updatedConnectionResp: UpdateConnectionResponse): void;
  onSaveFailed(err: unknown): void;
}

export default function DynamoDBForm(props: Props) {
  const { connectionId, defaultValues, onSaved, onSaveFailed } = props;
  const { account } = useAccount();
  const { mutateAsync: isConnectionNameAvailableAsync } = useMutation(
    isConnectionNameAvailable
  );
  const form = useForm<DynamoDbFormValues, EditConnectionFormContext>({
    resolver: yupResolver(DynamoDbFormValues),
    values: defaultValues,
    context: {
      originalConnectionName: defaultValues.connectionName,
      accountId: account?.id ?? '',
      isConnectionNameAvailable: isConnectionNameAvailableAsync,
    },
  });
  const { mutateAsync } = useMutation(updateConnection);

  async function onSubmit(values: DynamoDbFormValues) {
    try {
      const connectionResp = await mutateAsync(
        new UpdateConnectionRequest({
          id: connectionId,
          name: values.connectionName,
          connectionConfig: buildConnectionConfigDynamoDB(values),
        })
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

        <Accordion
          type="multiple"
          defaultValue={['credentials']}
          className="w-full"
        >
          <AccordionItem value="credentials">
            <AccordionTrigger>AWS Credentials</AccordionTrigger>
            <AccordionContent className="space-y-4">
              <p className="text-sm">
                This section is used to configure authentication credentials to
                allow access to the DynamoDB tables and will be specific to how
                you wish Neosync to connect to Dynamo.
              </p>
              <div className="space-y-8">
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
            <AccordionContent className="space-y-4">
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

        <div className="flex flex-row gap-3 justify-end">
          <Button type="submit">
            <ButtonText
              leftIcon={form.formState.isSubmitting ? <Spinner /> : <div></div>}
              text="Update"
            />
          </Button>
        </div>
      </form>
    </Form>
  );
}
