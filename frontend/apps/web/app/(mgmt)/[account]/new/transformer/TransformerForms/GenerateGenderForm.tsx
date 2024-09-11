'use client';
import { FormDescription, FormLabel } from '@/components/ui/form';

import FormErrorMessage from '@/components/FormErrorMessage';
import { Switch } from '@/components/ui/switch';
import { PlainMessage } from '@bufbuild/protobuf';
import { GenerateGender } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props
  extends TransformerConfigProps<
    GenerateGender,
    PlainMessage<GenerateGender>
  > {}

export default function GenerateGenderForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-col w-full">
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
        <FormErrorMessage message={errors?.abbreviate?.message} />
      </div>
    </div>
  );
}
