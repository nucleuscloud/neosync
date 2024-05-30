'use client';
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import {
  generateEmailTypeStringToEnum,
  getGenerateEmailTypeString,
  getInvalidEmailActionString,
  invalidEmailActionStringToEnum,
} from '@/util/util';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  GenerateEmailType,
  InvalidEmailAction,
  TransformEmail,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { TRANSFORMER_SCHEMA_CONFIGS } from '../../new/transformer/schema';
import { TransformerFormProps } from './util';

interface Props extends TransformerFormProps<TransformEmail> {}

export default function TransformEmailForm(props: Props): ReactElement {
  const { existingConfig, onSubmit, isReadonly } = props;

  const emailType =
    (existingConfig?.toJson() as any)?.emailType ?? // eslint-disable-line @typescript-eslint/no-explicit-any
    'GENERATE_EMAIL_TYPE_UUID_V4';
  const invalidEmailAction =
    (existingConfig?.toJson() as any)?.invalidEmailAction ?? // eslint-disable-line @typescript-eslint/no-explicit-any
    'INVALID_EMAIL_ACTION_REJECT';
  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver(TRANSFORMER_SCHEMA_CONFIGS.transformEmailConfig),
    defaultValues: {
      preserveDomain: existingConfig?.preserveDomain ?? false,
      preserveLength: existingConfig?.preserveLength ?? false,
      excludedDomains: existingConfig?.excludedDomains ?? [],
      emailType: emailType,
      invalidEmailAction: invalidEmailAction,
    },
  });

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <Form {...form}>
        <FormField
          control={form.control}
          name={`preserveLength`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Preserve Length</FormLabel>
                <FormDescription className="w-[90%]">
                  Set the length of the output email to be the same as the input
                </FormDescription>
              </div>
              <FormControl>
                <Switch
                  name={field.name}
                  checked={field.value}
                  disabled={isReadonly}
                  onCheckedChange={field.onChange}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={`preserveDomain`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Preserve Domain</FormLabel>
                <FormDescription className="w-[90%]">
                  Preserve the input domain including top level domain to the
                  output value. For ex. if the input is john@gmail.com, the
                  output will be ij23o@gmail.com
                </FormDescription>
              </div>
              <FormControl>
                <Switch
                  name={field.name}
                  checked={field.value}
                  disabled={isReadonly}
                  onCheckedChange={field.onChange}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={`excludedDomains`}
          render={({ field }) => (
            <FormItem className="rounded-lg border p-3 shadow-sm">
              <div className="flex flex-row items-start justify-between">
                <div className="flex flex-col space-y-2">
                  <FormLabel>Excluded Domains</FormLabel>
                  <FormDescription>
                    Provide a list of comma-separated domains that you want to
                    be excluded from the transformer. Do not provide an @ with
                    the domains.{' '}
                  </FormDescription>
                </div>
                <FormControl>
                  <div className="flex flex-col items-center">
                    <Input
                      {...field}
                      onChange={(e) =>
                        field.onChange(e.target.value.split(','))
                      }
                      disabled={isReadonly}
                      type="string"
                      className="min-w-[300px]"
                    />
                    <FormMessage />
                  </div>
                </FormControl>
              </div>
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={`emailType`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm gap-4 ">
              <div className="space-y-0.5">
                <FormLabel>Email Type</FormLabel>
                <FormDescription className="w-[90%]">
                  Configure the email type that will be used during
                  transformation.
                </FormDescription>
              </div>
              <FormControl>
                <Select
                  disabled={isReadonly}
                  onValueChange={(value) => {
                    // this is so hacky, but has to be done due to have we are encoding the incoming config and how the enums are converted to their wire-format string type
                    const emailConfig = new TransformEmail({
                      emailType: parseInt(value, 10),
                    }).toJson();
                    field.onChange((emailConfig as any).emailType); // eslint-disable-line @typescript-eslint/no-explicit-any
                  }}
                  value={generateEmailTypeStringToEnum(field.value).toString()}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {[
                      GenerateEmailType.UUID_V4,
                      GenerateEmailType.FULLNAME,
                    ].map((emailType) => (
                      <SelectItem
                        key={emailType}
                        className="cursor-pointer"
                        value={emailType.toString()}
                      >
                        {getGenerateEmailTypeString(emailType)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={`invalidEmailAction`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm gap-4 ">
              <div className="space-y-0.5">
                <FormLabel>Invalid Email Action</FormLabel>
                <FormDescription className="w-[90%]">
                  Configure the invalid email action that will be run in the
                  event the system encounters an email that does not conform to
                  RFC 5322.
                </FormDescription>
              </div>
              <FormControl>
                <Select
                  disabled={isReadonly}
                  onValueChange={(value) => {
                    // this is so hacky, but has to be done due to have we are encoding the incoming config and how the enums are converted to their wire-format string type
                    const emailConfig = new TransformEmail({
                      invalidEmailAction: parseInt(value, 10),
                    }).toJson();
                    field.onChange((emailConfig as any).invalidEmailAction); // eslint-disable-line @typescript-eslint/no-explicit-any
                  }}
                  value={invalidEmailActionStringToEnum(field.value).toString()}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {[
                      InvalidEmailAction.GENERATE,
                      InvalidEmailAction.NULL,
                      InvalidEmailAction.PASSTHROUGH,
                      InvalidEmailAction.REJECT,
                    ].map((action) => (
                      <SelectItem
                        key={action}
                        className="cursor-pointer"
                        value={action.toString()}
                      >
                        {getInvalidEmailActionString(action)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className="flex justify-end">
          <Button
            type="button"
            disabled={isReadonly}
            onClick={(e) => {
              form.handleSubmit((values) => {
                onSubmit(TransformEmail.fromJson(values));
              })(e);
            }}
          >
            Save
          </Button>
        </div>
      </Form>
    </div>
  );
}
