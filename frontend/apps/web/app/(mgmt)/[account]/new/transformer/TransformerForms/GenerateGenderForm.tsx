'use client';
import { FormDescription, FormLabel } from '@/components/ui/form';

import FormErrorMessage from '@/components/FormErrorMessage';
import { Switch } from '@/components/ui/switch';
import { create } from '@bufbuild/protobuf';
import { GenerateGender, GenerateGenderSchema } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<GenerateGender> {}

export default function GenerateGenderForm(props: Props): ReactElement<any> {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
      <div className="space-y-0.5 w-[80%]">
        <FormLabel>Abbreviate</FormLabel>
        <FormDescription>
          Abbreviate the gender to a single character. For example, female would
          be returned as f.
        </FormDescription>
      </div>
      <div className="flex flex-col">
        <div className="justify-end flex">
          <Switch
            checked={value.abbreviate}
            onCheckedChange={(checked) =>
              setValue(
                create(GenerateGenderSchema, {
                  ...value,
                  abbreviate: checked,
                })
              )
            }
            disabled={isDisabled}
          />
        </div>
        <FormErrorMessage message={errors?.abbreviate?.message} />
      </div>
    </div>
  );
}
