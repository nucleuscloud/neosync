'use client';
import FormErrorMessage from '@/components/FormErrorMessage';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { create } from '@bufbuild/protobuf';
import { TransformFloat64, TransformFloat64Schema } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<TransformFloat64> {}

export default function TransformFloat64Form(props: Props): ReactElement<any> {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-col w-full space-y-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Relative Minimum Range Value</FormLabel>
          <FormDescription>
            Sets a relative minium lower range value. This will create a
            lowerbound around the source input value. For example, if the input
            value is 10, and you set this value to 5, then the minimum range
            will be 5 (10-5 = 5).
          </FormDescription>
        </div>
        <div className="flex flex-col">
          <div className="justify-end flex">
            <div className="w-[300px]">
              <Input
                value={value.randomizationRangeMin}
                type="number"
                onChange={(e) => {
                  if (!isNaN(e.target.valueAsNumber)) {
                    setValue(
                      create(TransformFloat64Schema, {
                        ...value,
                        randomizationRangeMin: e.target.valueAsNumber,
                      })
                    );
                  }
                }}
                disabled={isDisabled}
              />
            </div>
          </div>
          <FormErrorMessage message={errors?.randomizationRangeMin?.message} />
        </div>
      </div>
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Relative Maximum Range Value</FormLabel>
          <FormDescription>
            Sets a relative maximum upper range value. This will create an
            upperbound around the source input value. For example, if the input
            value is 10, and you set this value to 5, then the maximum range
            will be 15 ( 10 + 5 = 15).
          </FormDescription>
        </div>
        <div className="flex flex-col">
          <div className="justify-end flex">
            <div className="w-[300px]">
              <Input
                value={value.randomizationRangeMax}
                type="number"
                onChange={(e) => {
                  if (!isNaN(e.target.valueAsNumber)) {
                    setValue(
                      create(TransformFloat64Schema, {
                        ...value,
                        randomizationRangeMax: e.target.valueAsNumber,
                      })
                    );
                  }
                }}
                disabled={isDisabled}
              />
            </div>
          </div>
          <FormErrorMessage message={errors?.randomizationRangeMax?.message} />
        </div>
      </div>
    </div>
  );
}
