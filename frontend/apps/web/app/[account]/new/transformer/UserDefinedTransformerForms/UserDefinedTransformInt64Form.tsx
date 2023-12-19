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
import { ReactElement } from 'react';
import { useFormContext } from 'react-hook-form';

interface Props {
  isDisabled?: boolean;
}

export default function UserDefinedTransformInt64Form(
  props: Props
): ReactElement {
  const fc = useFormContext();
  const { isDisabled } = props;
  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`config.config.value.randomizationRangeMin`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Minimum Range Value</FormLabel>
              <FormDescription className="w-[90%]">
                Sets a minium lower range value. This will create an lowerbound
                around the source input value. For example, if the input value
                is 10, and you set this value to 5, then the maximum range will
                be 5.
              </FormDescription>
            </div>
            <div className="flex flex-col h-14">
              <div className="justify-end flex">
                <FormControl>
                  <div className="max-w-[180px]">
                    <Input
                      value={field.value !== null ? String(field.value) : ''}
                      type="number"
                      onChange={(e) => {
                        field.onChange(
                          e.target.value === '' ? null : Number(e.target.value)
                        );
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
        name={`config.config.value.randomizationRangeMax`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Maxiumum Range Value</FormLabel>
              <FormDescription className="w-[90%]">
                Sets a maximum upper range value. This will create an upperbound
                around the source input value. For example, if the input value
                is 10, and you set this value to 5, then the maximum range will
                be 15.
              </FormDescription>
            </div>
            <div className="flex flex-col h-14">
              <div className="justify-end flex">
                <FormControl>
                  <div className="max-w-[180px]">
                    <Input
                      value={field.value !== null ? String(field.value) : ''}
                      type="number"
                      onChange={(e) => {
                        field.onChange(
                          e.target.value === '' ? null : Number(e.target.value)
                        );
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
