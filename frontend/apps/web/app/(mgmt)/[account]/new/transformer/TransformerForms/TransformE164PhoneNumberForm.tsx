'use client';
import FormErrorMessage from '@/components/FormErrorMessage';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Switch } from '@/components/ui/switch';
import { create } from '@bufbuild/protobuf';
import {
  TransformE164PhoneNumber,
  TransformE164PhoneNumberSchema,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<TransformE164PhoneNumber> {}

export default function TransformE164NumberForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
      <div className="space-y-0.5 w-[80%]">
        <FormLabel>Preserve Length</FormLabel>
        <FormDescription>
          Set the length of the output e164 phone number to be the same as the
          input e164 phone number.
        </FormDescription>
      </div>
      <div className="flex flex-col">
        <div className="justify-end flex">
          <Switch
            checked={value.preserveLength}
            onCheckedChange={(checked) =>
              setValue(
                create(TransformE164PhoneNumberSchema, {
                  preserveLength: checked,
                })
              )
            }
            disabled={isDisabled}
          />
        </div>
        <FormErrorMessage message={errors?.preserveLength?.message} />
      </div>
    </div>
  );
}
