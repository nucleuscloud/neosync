'use client';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Switch } from '@/components/ui/switch';
import { TransformFullName } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<TransformFullName> {}

export default function UserDefinedTransformFullNameForm(
  props: Props
): ReactElement {
  const { value, setValue, isDisabled } = props;
  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5">
          <FormLabel>Preserve Length</FormLabel>
          <FormDescription className="w-[90%]">
            Generates a full name which has the same first name and last name
            length as the input first and last names
          </FormDescription>
        </div>
        <Switch
          checked={value.preserveLength}
          onCheckedChange={(checked) =>
            setValue(
              new TransformFullName({ ...value, preserveLength: checked })
            )
          }
          disabled={isDisabled}
        />
      </div>
    </div>
  );
}
