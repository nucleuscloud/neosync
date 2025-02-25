'use client';
import FormErrorMessage from '@/components/FormErrorMessage';
import { FormDescription, FormLabel } from '@/components/ui/form';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { getGenerateEmailTypeString } from '@/util/util';
import { create } from '@bufbuild/protobuf';
import {
  GenerateEmail,
  GenerateEmailSchema,
  GenerateEmailType,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<GenerateEmail> {}

export default function GenerateEmailForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
      <div className="space-y-0.5 w-[80%]">
        <FormLabel>Email Type</FormLabel>
        <FormDescription>
          Select the type of email you want to generate. Uuid_v4 emails
          guarantee uniqueness.
        </FormDescription>
      </div>
      <div className="flex flex-col">
        <div className="justify-end flex">
          <Select
            disabled={isDisabled}
            onValueChange={(newValue) => {
              setValue(
                create(GenerateEmailSchema, {
                  ...value,
                  // this is so hacky, but has to be done due to have we are encoding the incoming config and how the enums are converted to their wire-format string type
                  emailType: parseInt(newValue, 10),
                })
              );
            }}
            value={value.emailType?.toString()}
          >
            <SelectTrigger className="w-[300px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {[GenerateEmailType.UUID_V4, GenerateEmailType.FULLNAME].map(
                (emailType) => (
                  <SelectItem
                    key={emailType}
                    className="cursor-pointer"
                    value={emailType.toString()}
                  >
                    {getGenerateEmailTypeString(emailType)}
                  </SelectItem>
                )
              )}
            </SelectContent>
          </Select>
        </div>
        <FormErrorMessage message={errors?.emailType?.message} />
      </div>
    </div>
  );
}
