'use client';
import FormErrorMessage from '@/components/FormErrorMessage';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Switch } from '@/components/ui/switch';
import { PlainMessage } from '@bufbuild/protobuf';
import { GenerateUuid } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props
  extends TransformerConfigProps<GenerateUuid, PlainMessage<GenerateUuid>> {}

export default function GenerateUuidForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-col w-full space-y-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Include hyphens</FormLabel>
          <FormDescription>
            Set to true to include hyphens in the generated UUID. Note: some
            databases such as Postgres automatically convert UUIDs with no
            hyphens to have hyphens when they store the data.
          </FormDescription>
        </div>
        <Switch
          checked={value.includeHyphens}
          onCheckedChange={(checked) =>
            setValue(new GenerateUuid({ ...value, includeHyphens: checked }))
          }
          disabled={isDisabled}
        />
      </div>
      <FormErrorMessage message={errors?.includeHyphens?.message} />
    </div>
  );
}
