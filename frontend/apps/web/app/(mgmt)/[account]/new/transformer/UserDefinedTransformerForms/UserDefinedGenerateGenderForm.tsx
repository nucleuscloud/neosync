'use client';
import { FormDescription, FormLabel } from '@/components/ui/form';

import { Switch } from '@/components/ui/switch';
import { GenerateGender } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<GenerateGender> {}

export default function UserDefinedGenerateGenderForm(
  props: Props
): ReactElement {
  const { value, setValue, isDisabled } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5">
          <FormLabel>Abbreviate</FormLabel>
          <FormDescription className="w-[90%]">
            Abbreviate the gender to a single character. For example, female
            would be returned as f.
          </FormDescription>
        </div>
        <Switch
          checked={value.abbreviate}
          onCheckedChange={(checked) =>
            setValue(new GenerateGender({ ...value, abbreviate: checked }))
          }
          disabled={isDisabled}
        />
      </div>
    </div>
  );
}
