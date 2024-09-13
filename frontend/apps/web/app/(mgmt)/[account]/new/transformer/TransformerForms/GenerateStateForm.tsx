'use client';
import { FormDescription, FormLabel } from '@/components/ui/form';

import FormErrorMessage from '@/components/FormErrorMessage';
import { Switch } from '@/components/ui/switch';
import { PlainMessage } from '@bufbuild/protobuf';
import { GenerateState } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props
  extends TransformerConfigProps<GenerateState, PlainMessage<GenerateState>> {}

export default function GenerateStateForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
      <div className="space-y-0.5 w-[80%]">
        <FormLabel>Generate Full Name</FormLabel>
        <FormDescription>
          Enable to return the full state name with a capitalized first letter.
          Returns the 2-letter state code by default.
        </FormDescription>
      </div>
      <div className="flex flex-col ">
        <div className="justify-end flex">
          <Switch
            checked={value.generateFullName}
            onCheckedChange={(checked) =>
              setValue(
                new GenerateState({ ...value, generateFullName: checked })
              )
            }
            disabled={isDisabled}
          />
        </div>
        <FormErrorMessage message={errors?.generateFullName?.message} />
      </div>
    </div>
  );
}
