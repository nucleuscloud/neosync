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
import { PlainMessage } from '@bufbuild/protobuf';
import { GenerateEmail, GenerateEmailType } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props
  extends TransformerConfigProps<GenerateEmail, PlainMessage<GenerateEmail>> {}

export default function GenerateEmailForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <div className="space-y-0.5">
        <FormLabel>Email Type</FormLabel>
        <FormDescription>
          Configure the email type that will be used during generation.
        </FormDescription>
      </div>
      <Select
        disabled={isDisabled}
        onValueChange={(newValue) => {
          setValue(
            new GenerateEmail({
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
      <FormErrorMessage message={errors?.emailType?.message} />
    </div>
  );
}
