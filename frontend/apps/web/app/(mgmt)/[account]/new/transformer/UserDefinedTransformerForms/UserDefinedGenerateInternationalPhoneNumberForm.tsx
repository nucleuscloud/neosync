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

export default function UserDefinedGenerateInternationalPhoneNumberForm(
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
        name={`config.value.min`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Minimum Length</FormLabel>
              <FormDescription>
                Set the minimum length range of the output phone number. It
                cannot be less than 9.
              </FormDescription>
            </div>
            <div className="flex flex-col h-14">
              <div className="justify-end flex">
                <FormControl>
                  <div className="w-[300px]">
                    <Input
                      value={field.value ? parseInt(field.value) : 0}
                      type="number"
                      onChange={(e) => {
                        if (!isNaN(e.target.valueAsNumber)) {
                          field.onChange(BigInt(e.target.valueAsNumber));
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
            <div className="space-y-0.5 ">
              <FormLabel>Maximum Length</FormLabel>
              <FormDescription>
                Set the maximum length range of the output phone number. It
                cannot be greater than 15.
              </FormDescription>
            </div>
            <div className="flex flex-col h-14">
              <div className="justify-end flex">
                <FormControl>
                  <div className="w-[300px]">
                    <Input
                      value={field.value ? parseInt(field.value) : 1}
                      type="number"
                      onChange={(e) => {
                        if (!isNaN(e.target.valueAsNumber)) {
                          field.onChange(BigInt(e.target.valueAsNumber));
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
