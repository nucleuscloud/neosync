'use client';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { ReactElement, useEffect } from 'react';
import { useFormContext } from 'react-hook-form';
import {
  CreateUserDefinedTransformerSchema,
  UpdateUserDefinedTransformer,
} from '../schema';

interface Props {
  isDisabled?: boolean;
}

export default function UserDefinedTransformFloat64Form(
  props: Props
): ReactElement {
  const fc = useFormContext<
    UpdateUserDefinedTransformer | CreateUserDefinedTransformerSchema
  >();
  const { isDisabled } = props;

  const min = fc.watch('config.value.randomizationRangeMin');
  const max = fc.watch('config.value.randomizationRangeMax');

  useEffect(() => {
    fc.trigger('config.value.randomizationRangeMin');
  }, [max, fc.trigger]);

  useEffect(() => {
    fc.trigger('config.value.randomizationRangeMax');
  }, [min, fc.trigger]);

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`config.value.randomizationRangeMin`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Relative Minimum Range Value</FormLabel>
              <FormDescription className="w-[90%]">
                Sets a relative minium lower range value. This will create a
                lowerbound around the source input value. For example, if the
                input value is 10, and you set this value to 5, then the minimum
                range will be 5 (10-5 = 5).
              </FormDescription>
            </div>
            <div className="flex flex-col h-14">
              <div className="justify-end flex">
                <FormControl>
                  <div className="w-[300px]">
                    <Input
                      value={field.value ? parseFloat(field.value) : 0}
                      type="number"
                      onChange={(e) => {
                        if (!isNaN(e.target.valueAsNumber)) {
                          field.onChange(e.target.valueAsNumber);
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
        )}
      />
      <FormField
        name={`config.value.randomizationRangeMax`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Relative Maximum Range Value</FormLabel>
              <FormDescription className="w-[90%]">
                Sets a relative maximum upper range value. This will create an
                upperbound around the source input value. For example, if the
                input value is 10, and you set this value to 5, then the maximum
                range will be 15 ( 10 + 5 = 15).
              </FormDescription>
            </div>
            <div className="flex flex-col h-14">
              <div className="justify-end flex">
                <FormControl>
                  <div className="w-[300px]">
                    <Input
                      value={field.value ? parseFloat(field.value) : 1}
                      type="number"
                      onChange={(e) => {
                        if (!isNaN(e.target.valueAsNumber)) {
                          field.onChange(e.target.valueAsNumber);
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
        )}
      />
    </div>
  );
}
