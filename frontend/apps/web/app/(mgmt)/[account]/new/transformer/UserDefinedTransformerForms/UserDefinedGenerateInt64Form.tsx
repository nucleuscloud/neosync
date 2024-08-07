'use client';
import { FormDescription, FormItem, FormLabel } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';

import { GenerateInt64 } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<GenerateInt64> {}

export default function UserDefinedGenerateInt64Form(
  props: Props
): ReactElement {
  const { value, setValue, isDisabled } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5">
          <FormLabel>Randomize Sign</FormLabel>
          <FormDescription>
            {`After the value has been generated, will randomly flip the sign. This may cause the generated value to be out of the defined min/max range.
                  If the min/max is 20-40, the value may be in the following ranges: 20 <= x <= 40 and -40 <= x <= -20`}
          </FormDescription>
        </div>
        <div className="flex flex-col h-14">
          <div className="justify-end flex">
            <div className="w-[300px]">
              <Switch
                checked={value.randomizeSign}
                onCheckedChange={(checked) =>
                  setValue(
                    new GenerateInt64({ ...value, randomizeSign: checked })
                  )
                }
                disabled={isDisabled}
              />
            </div>
          </div>
        </div>
      </div>
      <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5">
          <FormLabel>Minimum Value</FormLabel>
          <FormDescription>
            Sets a minimum range for generated int64 value.
          </FormDescription>
        </div>
        <div className="flex flex-col h-14">
          <div className="justify-end flex">
            <div className="w-[300px]">
              <Input
                type="number"
                value={value.min ? parseInt(value.min.toString(), 10) : 0}
                onChange={(e) => {
                  if (!isNaN(e.target.valueAsNumber)) {
                    setValue(
                      new GenerateInt64({
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
        </div>
      </FormItem>
      <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5">
          <FormLabel>Maximum Value</FormLabel>
          <FormDescription>
            Sets a maximum range for generated int64 value.
          </FormDescription>
        </div>
        <div className="flex flex-col h-14">
          <div className="justify-end flex">
            <div className="w-[300px]">
              <Input
                type="number"
                value={value.max ? parseInt(value.max.toString(), 10) : 1}
                onChange={(e) => {
                  if (!isNaN(e.target.valueAsNumber)) {
                    setValue(
                      new GenerateInt64({
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
        </div>
      </FormItem>
    </div>
  );
}
