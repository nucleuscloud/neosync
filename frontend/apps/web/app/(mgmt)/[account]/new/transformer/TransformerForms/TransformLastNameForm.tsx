'use client';
import FormErrorMessage from '@/components/FormErrorMessage';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Switch } from '@/components/ui/switch';
import { PlainMessage } from '@bufbuild/protobuf';
import { TransformLastName } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props
  extends TransformerConfigProps<
    TransformLastName,
    PlainMessage<TransformLastName>
  > {}

export default function TransformLastNameForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;
  return (
    <div className="flex flex-col w-full space-y-4 ">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Preserve Length</FormLabel>
          <FormDescription>
            Set the length of the output last name to be the same as the input
          </FormDescription>
        </div>
        <Switch
          checked={value.preserveLength}
          onCheckedChange={(checked) =>
            setValue(
              new TransformLastName({ ...value, preserveLength: checked })
            )
          }
          disabled={isDisabled}
        />
      </div>
      <FormErrorMessage message={errors?.preserveLength?.message} />
    </div>
  );
}
