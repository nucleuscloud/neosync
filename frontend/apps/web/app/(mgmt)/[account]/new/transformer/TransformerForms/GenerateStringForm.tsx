'use client';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Input } from '@/components/ui/input';

import FormErrorMessage from '@/components/FormErrorMessage';
import { create } from '@bufbuild/protobuf';
import { GenerateString, GenerateStringSchema } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<GenerateString> {}

export default function GenerateStringForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-col w-full space-y-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Minimum Length</FormLabel>
          <FormDescription>
            Set the minimum length range of the output string.
          </FormDescription>
        </div>
        <div className="flex flex-col">
          <div className="justify-end flex">
            <div className="max-w-[300px]">
              <Input
                value={value.min ? parseInt(value.min.toString(), 10) : 0}
                type="number"
                onChange={(e) => {
                  if (!isNaN(e.target.valueAsNumber)) {
                    setValue(
                      create(GenerateStringSchema, {
                        ...value,
                        min: BigInt(e.target.valueAsNumber),
                      })
                    );
                  }
                }}
                disabled={isDisabled}
              />
            </div>
          </div>
          <FormErrorMessage message={errors?.min?.message} />
        </div>
      </div>
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Maximum Length</FormLabel>
          <FormDescription>
            Set the maximum length range of the output string.
          </FormDescription>
        </div>
        <div className="flex flex-col">
          <div className="justify-end flex">
            <div className="max-w-[300px]">
              <Input
                value={value.max ? parseInt(value.max.toString()) : 0}
                type="number"
                onChange={(e) => {
                  if (!isNaN(e.target.valueAsNumber)) {
                    setValue(
                      create(GenerateStringSchema, {
                        ...value,
                        max: BigInt(e.target.valueAsNumber),
                      })
                    );
                  }
                }}
                disabled={isDisabled}
              />
            </div>
          </div>
          <FormErrorMessage message={errors?.max?.message} />
        </div>
      </div>
    </div>
  );
}
