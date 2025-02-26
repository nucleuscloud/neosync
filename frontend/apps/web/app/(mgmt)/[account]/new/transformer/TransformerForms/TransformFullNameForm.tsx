'use client';
import FormErrorMessage from '@/components/FormErrorMessage';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Switch } from '@/components/ui/switch';
import { create } from '@bufbuild/protobuf';
import { TransformFullName, TransformFullNameSchema } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<TransformFullName> {}

export default function TransformFullNameForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;
  return (
    <div className="flex flex-col w-full space-y-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Preserve Length</FormLabel>
          <FormDescription>
            Generates a full name which has the same first name and last name
            length as the input first and last names
          </FormDescription>
        </div>
        <Switch
          checked={value.preserveLength}
          onCheckedChange={(checked) =>
            setValue(
              create(TransformFullNameSchema, {
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
