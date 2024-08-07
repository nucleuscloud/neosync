'use client';
import {
  FormControl,
  FormDescription,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';

import { GenerateFloat64 } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<GenerateFloat64> {}

export default function UserDefinedGenerateFloat64Form(
  props: Props
): ReactElement {
  const { value, setValue, isDisabled } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5">
          <FormLabel>Randomize Sign</FormLabel>
          <FormDescription className="w-[80%]">
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
                    new GenerateFloat64({ ...value, randomizeSign: checked })
                  )
                }
                disabled={isDisabled}
              />
            </div>
          </div>
        </div>
      </div>
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5">
          <FormLabel>Minimum Value</FormLabel>
          <FormDescription>
            Sets a minimum range for generated float64 value.
          </FormDescription>
        </div>
        <div className="flex flex-col h-14">
          <div className="justify-end flex">
            <div className="w-[300px]">
              <Input
                value={value.min}
                type="number"
                onChange={(e) => {
                  if (!isNaN(e.target.valueAsNumber)) {
                    setValue(
                      new GenerateFloat64({
                        ...value,
                        min: e.target.valueAsNumber,
                      })
                    );
                  }
                }}
                disabled={isDisabled}
              />
            </div>
          </div>
        </div>
      </div>
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5">
          <FormLabel>Maximum Value</FormLabel>
          <FormDescription>
            Sets a maximum range for generated float64 value.
          </FormDescription>
        </div>
        <div className="flex flex-col h-14">
          <div className="justify-end flex">
            <div className="w-[300px]">
              <Input
                value={value.max}
                type="number"
                onChange={(e) => {
                  if (!isNaN(e.target.valueAsNumber)) {
                    setValue(
                      new GenerateFloat64({
                        ...value,
                        max: e.target.valueAsNumber,
                      })
                    );
                  }
                }}
                disabled={isDisabled}
              />
            </div>
          </div>
        </div>
      </div>
      <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5">
          <FormLabel>Precision</FormLabel>
          <FormDescription>
            Sets the precision for the entire float64 value, not just the
            decimals. For example. a precision of 4 would update a float64 value
            of 23.567 to 23.56.
          </FormDescription>
        </div>
        <div className="flex flex-col h-14">
          <div className="justify-end flex">
            <FormControl>
              <div className="w-[300px]">
                <Input
                  type="number"
                  value={
                    value.precision
                      ? parseInt(value.precision.toString(), 10)
                      : 1
                  }
                  onChange={(e) => {
                    if (!isNaN(e.target.valueAsNumber)) {
                      setValue(
                        new GenerateFloat64({
                          ...value,
                          precision: BigInt(e.target.valueAsNumber),
                        })
                      );
                    }
                  }}
                  disabled={isDisabled}
                />
              </div>
            </FormControl>
          </div>
          <FormMessage />
        </div>
      </FormItem>
    </div>
  );
}
