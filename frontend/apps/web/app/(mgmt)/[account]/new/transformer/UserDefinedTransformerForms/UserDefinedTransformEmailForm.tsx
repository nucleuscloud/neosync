'use client';
import {
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
import {
  GenerateEmailType,
  InvalidEmailAction,
  TransformEmail,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { useFormContext } from 'react-hook-form';
import {
  CreateUserDefinedTransformerSchema,
  UpdateUserDefinedTransformer,
} from '../schema';

interface Props {
  isDisabled?: boolean;
}

export default function UserDefinedTransformEmailForm(
  props: Props
): ReactElement {
  const fc = useFormContext<
    UpdateUserDefinedTransformer | CreateUserDefinedTransformerSchema
  >();
  const { isDisabled } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`config.value.preserveLength`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Preserve Length</FormLabel>
              <FormDescription className="w-[90%]">
                Set the length of the output email to be the same as the input
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={field.value}
                onCheckedChange={field.onChange}
                disabled={isDisabled}
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
      <FormField
        name={`config.value.preserveDomain`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Preserve Domain</FormLabel>
              <FormDescription className="w-[90%]">
                Preserve the input domain including top level domain to the
                output value. For ex. if the input is john@gmail.com, the output
                will be ij23o@gmail.com
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={field.value}
                onCheckedChange={field.onChange}
                disabled={isDisabled}
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
      <FormField
        name={`config.value.excludedDomains`}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Excluded Domains</FormLabel>
              <FormDescription>
                Provide a list of comma-separated domains that you want to be
                excluded from the transformer. Do not provide an @ with the
                domains.{' '}
              </FormDescription>
            </div>
            <FormControl>
              <div className="min-w-[300px]">
                <Input
                  {...field}
                  type="string"
                  className="min-w-[300px]"
                  onChange={(e) => field.onChange(e.target.value.split(','))}
                  disabled={isDisabled}
                />
              </div>
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
      <FormField
        name={`config.value.emailType`}
        control={fc.control}
        render={({ field }) => {
          return (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Email Type</FormLabel>
                <FormDescription>
                  Configure the email type that will be used during
                  transformation.
                </FormDescription>
              </div>
              <FormControl>
                <Select
                  disabled={isDisabled}
                  onValueChange={(value) => {
                    // this is so hacky, but has to be done due to have we are encoding the incoming config and how the enums are converted to their wire-format string type
                    const emailConfig = new TransformEmail({
                      emailType: parseInt(value, 10),
                    }).toJson();
                    field.onChange((emailConfig as any).emailType); // eslint-disable-line @typescript-eslint/no-explicit-any
                  }}
                  value={generateEmailTypeStringToEnum(field.value).toString()}
                >
                  <SelectTrigger className="w-[300px]">
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
          );
        }}
      />
      <FormField
        name={`config.value.invalidEmailAction`}
        control={fc.control}
        render={({ field }) => {
          return (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Invalid Email Action</FormLabel>
                <FormDescription>
                  Configure the invalid email action that will be run in the
                  event the system encounters an email that does not conform to
                  RFC 5322.
                </FormDescription>
              </div>
              <FormControl>
                <Select
                  disabled={isDisabled}
                  onValueChange={(value) => {
                    // this is so hacky, but has to be done due to have we are encoding the incoming config and how the enums are converted to their wire-format string type
                    const emailConfig = new TransformEmail({
                      invalidEmailAction: parseInt(value, 10),
                    }).toJson();
                    field.onChange((emailConfig as any).invalidEmailAction); // eslint-disable-line @typescript-eslint/no-explicit-any
                  }}
                  value={invalidEmailActionStringToEnum(field.value).toString()}
                >
                  <SelectTrigger className="w-[300px]">
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
          );
        }}
      />
    </div>
  );
}
