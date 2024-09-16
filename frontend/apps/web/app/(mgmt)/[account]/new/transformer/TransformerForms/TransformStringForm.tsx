'use client';
import { FormDescription, FormLabel } from '@/components/ui/form';

import FormErrorMessage from '@/components/FormErrorMessage';
import { Switch } from '@/components/ui/switch';
import { PlainMessage } from '@bufbuild/protobuf';
import { TransformString } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props
  extends TransformerConfigProps<
    TransformString,
    PlainMessage<TransformString>
  > {}

export default function TransformStringForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-col w-full space-y-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Preserve Length</FormLabel>
          <FormDescription>
            Set the length of the output string to be the same as the input
          </FormDescription>
        </div>
        <Switch
          checked={value.preserveLength}
          onCheckedChange={(checked) =>
            setValue(new TransformString({ ...value, preserveLength: checked }))
          }
          disabled={isDisabled}
        />
      </div>
      <FormErrorMessage message={errors?.preserveLength?.message} />
    </div>
  );
}
