'use client';
import FormErrorMessage from '@/components/FormErrorMessage';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Switch } from '@/components/ui/switch';
import { create } from '@bufbuild/protobuf';
import {
  TransformInt64PhoneNumber,
  TransformInt64PhoneNumberSchema,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<TransformInt64PhoneNumber> {}

export default function TransformIntPhoneNumberForm(
  props: Props
): ReactElement<any> {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-col w-full space-y-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700  p-3 shadow-xs">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Preserve Length</FormLabel>
          <FormDescription>
            Set the length of the output phone number to be the same as the
            input
          </FormDescription>
        </div>
        <Switch
          checked={value.preserveLength}
          onCheckedChange={(checked) =>
            setValue(
              create(TransformInt64PhoneNumberSchema, {
                ...value,
                preserveLength: checked,
              })
            )
          }
          disabled={isDisabled}
        />
      </div>
      <FormErrorMessage message={errors?.preserveLength?.message} />
    </div>
  );
}
