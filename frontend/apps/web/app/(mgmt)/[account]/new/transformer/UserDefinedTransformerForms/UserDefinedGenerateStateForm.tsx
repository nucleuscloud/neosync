'use client';
import { FormDescription, FormLabel } from '@/components/ui/form';

import { Switch } from '@/components/ui/switch';
import { GenerateState } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<GenerateState> {}

export default function UserDefinedGenerateStateForm(
  props: Props
): ReactElement {
  const { value, setValue, isDisabled } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5">
          <FormLabel>Generate Full Name</FormLabel>
          <FormDescription>
            Set to true to return the full state name with a capitalized first
            letter. Returns the 2-letter state code by default.
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
        </div>
      </div>
    </div>
  );
}
