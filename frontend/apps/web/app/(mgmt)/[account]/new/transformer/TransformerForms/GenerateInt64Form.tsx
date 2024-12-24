'use client';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';

import FormErrorMessage from '@/components/FormErrorMessage';
import { create } from '@bufbuild/protobuf';
import { GenerateInt64, GenerateInt64Schema } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<GenerateInt64> {}

export default function GenerateInt64Form(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-col w-full space-y-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Randomize Sign</FormLabel>
          <FormDescription>
            {`Will randomly assign the sign.  This may cause the generated value to be out of the defined min/max range.
                  If the min/max is 20-40, the value may be in the following ranges: 20 <= x <= 40 and -40 <= x <= -20`}
          </FormDescription>
        </div>
        <div className="flex flex-col">
          <div className="justify-end flex">
            <Switch
              checked={value.randomizeSign}
              onCheckedChange={(checked) =>
                setValue(
                  create(GenerateInt64Schema, {
                    ...value,
                    randomizeSign: checked,
                  })
                )
              }
              disabled={isDisabled}
            />
          </div>
          <FormErrorMessage message={errors?.randomizeSign?.message} />
        </div>
      </div>
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5">
          <FormLabel>Minimum Value</FormLabel>
          <FormDescription>
            Sets a minimum range for generated int64 value.
          </FormDescription>
        </div>
        <div className="flex flex-col">
          <div className="justify-end flex">
            <div className="w-[300px]">
              <Input
                type="number"
                value={value.min ? parseInt(value.min.toString(), 10) : 0}
                onChange={(e) => {
                  if (!isNaN(e.target.valueAsNumber)) {
                    setValue(
                      create(GenerateInt64Schema, {
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
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5">
          <FormLabel>Maximum Value</FormLabel>
          <FormDescription>
            Sets a maximum range for generated int64 value.
          </FormDescription>
        </div>
        <div className="flex flex-col">
          <div className="justify-end flex">
            <div className="w-[300px]">
              <Input
                type="number"
                value={value.max ? parseInt(value.max.toString(), 10) : 0}
                onChange={(e) => {
                  if (!isNaN(e.target.valueAsNumber)) {
                    setValue(
                      create(GenerateInt64Schema, {
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
