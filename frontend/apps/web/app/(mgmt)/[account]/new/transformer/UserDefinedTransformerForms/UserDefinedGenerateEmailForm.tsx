'use client';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  generateEmailTypeStringToEnum,
  getGenerateEmailTypeString,
} from '@/util/util';
import { GenerateEmail, GenerateEmailType } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useFormContext } from 'react-hook-form';
import {
  CreateUserDefinedTransformerSchema,
  UpdateUserDefinedTransformer,
} from '../schema';

interface Props {
  isDisabled?: boolean;
}

export default function UserDefinedGenerateEmailForm(
  props: Props
): ReactElement {
  const fc = useFormContext<
    UpdateUserDefinedTransformer | CreateUserDefinedTransformerSchema
  >();

  const { isDisabled } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`config.value.emailType`}
        control={fc.control}
        render={({ field }) => {
          return (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Email Type</FormLabel>
                <FormDescription>
                  Configure the email type that will be used during generation.
                </FormDescription>
              </div>
              <FormControl>
                <Select
                  disabled={isDisabled}
                  onValueChange={(value) => {
                    // this is so hacky, but has to be done due to have we are encoding the incoming config and how the enums are converted to their wire-format string type
                    const emailConfig = new GenerateEmail({
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
    </div>
  );
}
