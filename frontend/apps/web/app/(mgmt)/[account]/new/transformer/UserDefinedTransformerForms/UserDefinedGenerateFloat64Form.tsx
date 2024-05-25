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
import { Switch } from '@/components/ui/switch';

import { ReactElement, useEffect } from 'react';
import { useFormContext } from 'react-hook-form';
import {
  CreateUserDefinedTransformerSchema,
  UpdateUserDefinedTransformer,
} from '../schema';

interface Props {
  isDisabled?: boolean;
}

export default function UserDefinedGenerateFloat64Form(
  props: Props
): ReactElement {
  const fc = useFormContext<
    UpdateUserDefinedTransformer | CreateUserDefinedTransformerSchema
  >();

  const { isDisabled } = props;

  const min = fc.watch('config.value.min');
  const max = fc.watch('config.value.max');

  useEffect(() => {
    fc.trigger('config.value.min');
  }, [max, fc.trigger]);

  useEffect(() => {
    fc.trigger('config.value.max');
  }, [min, fc.trigger]);

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`config.value.randomizeSign`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Randomize Sign</FormLabel>
              <FormDescription className="w-[80%]">
                {`After the value has been generated, will randomly flip the sign. This may cause the generated value to be out of the defined min/max range.
                  If the min/max is 20-40, the value may be in the following ranges: 20 <= x <= 40 and -40 <= x <= -20`}
              </FormDescription>
            </div>
            <div className="flex flex-col h-14">
              <div className="justify-end flex">
                <FormControl>
                  <div className="w-[300px]">
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
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
        name={`config.value.min`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Minimum Value</FormLabel>
              <FormDescription>
                Sets a minimum range for generated float64 value.
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
        name={`config.value.max`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Maximum Value</FormLabel>
              <FormDescription>
                Sets a maximum range for generated float64 value.
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
      <FormField
        name={`config.value.precision`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Precision</FormLabel>
              <FormDescription>
                Sets the precision for the entire float64 value, not just the
                decimals. For example. a precision of 4 would update a float64
                value of 23.567 to 23.56.
              </FormDescription>
            </div>
            <div className="flex flex-col h-14">
              <div className="justify-end flex">
                <FormControl>
                  <div className="w-[300px]">
                    <Input
                      type="number"
                      value={field.value ? parseInt(field.value) : 1}
                      onChange={(e) => {
                        if (!isNaN(e.target.valueAsNumber)) {
                          field.onChange(Number(e.target.valueAsNumber));
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
