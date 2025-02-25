'use client';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';

import FormErrorMessage from '@/components/FormErrorMessage';
import { create } from '@bufbuild/protobuf';
import { GenerateFloat64, GenerateFloat64Schema } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<GenerateFloat64> {}

export default function GenerateFloat64Form(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-col w-full space-y-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Randomize Sign</FormLabel>
          <FormDescription>
            {`Will randomly assign the sign. This may cause the generated value to be out of the defined min/max range.
                  If the min/max is 20-40, the value may be in the following ranges: 20 <= x <= 40 and -40 <= x <= -20`}
          </FormDescription>
        </div>
        <div className="flex flex-col">
          <div className="justify-end flex">
            <Switch
              checked={value.randomizeSign}
              onCheckedChange={(checked) =>
                setValue(
                  create(GenerateFloat64Schema, {
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
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Minimum Value</FormLabel>
          <FormDescription>
            Sets a minimum range for generated float64 value.
          </FormDescription>
        </div>
        <div className="flex flex-col">
          <div className="justify-end flex">
            <div className="w-[300px]">
              <Input
                value={value.min}
                type="number"
                onChange={(e) => {
                  if (!isNaN(e.target.valueAsNumber)) {
                    setValue(
                      create(GenerateFloat64Schema, {
                        ...value,
                        min: e.target.valueAsNumber,
                      })
                    );
                  }
                }}
                disabled={isDisabled}
              />
              <FormErrorMessage message={errors?.min?.message} />
            </div>
          </div>
        </div>
      </div>
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
        <div className="space-y-0.5">
          <FormLabel>Maximum Value</FormLabel>
          <FormDescription>
            Sets a maximum range for generated float64 value.
          </FormDescription>
        </div>
        <div className="flex flex-col">
          <div className="justify-end flex">
            <div className="w-[300px]">
              <Input
                value={value.max}
                type="number"
                onChange={(e) => {
                  if (!isNaN(e.target.valueAsNumber)) {
                    setValue(
                      create(GenerateFloat64Schema, {
                        ...value,
                        max: e.target.valueAsNumber,
                      })
                    );
                  }
                }}
                disabled={isDisabled}
              />
              <FormErrorMessage message={errors?.max?.message} />
            </div>
          </div>
        </div>
      </div>
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
        <div className="space-y-0.5">
          <FormLabel>Precision</FormLabel>
          <FormDescription>
            Sets the precision for the entire float64 value, not just the
            decimals. For example. a precision of 4 would update a float64 value
            of 23.567 to 23.56.
          </FormDescription>
        </div>
        <div className="flex flex-col ">
          <div className="justify-end flex">
            <div className="w-[300px]">
              <Input
                type="number"
                value={
                  value.precision ? parseInt(value.precision.toString(), 10) : 0
                }
                onChange={(e) => {
                  if (!isNaN(e.target.valueAsNumber)) {
                    setValue(
                      create(GenerateFloat64Schema, {
                        ...value,
                        precision: BigInt(e.target.valueAsNumber),
                      })
                    );
                  }
                }}
                disabled={isDisabled}
              />
            </div>
          </div>
          <FormErrorMessage message={errors?.precision?.message} />
        </div>
      </div>
    </div>
  );
}
