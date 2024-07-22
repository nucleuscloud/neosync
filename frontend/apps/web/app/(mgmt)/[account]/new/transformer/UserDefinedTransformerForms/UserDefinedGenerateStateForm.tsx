'use client';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';

import { Switch } from '@/components/ui/switch';
import { ReactElement } from 'react';
import { useFormContext } from 'react-hook-form';
import {
  CreateUserDefinedTransformerFormValues,
  UpdateUserDefinedTransformerFormValues,
} from '../schema';

interface Props {
  isDisabled?: boolean;
}

export default function UserDefinedGenerateStateForm(
  props: Props
): ReactElement {
  const fc = useFormContext<
    | UpdateUserDefinedTransformerFormValues
    | CreateUserDefinedTransformerFormValues
  >();

  const { isDisabled } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`config.value.generateFullName`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Generate Full Name</FormLabel>
              <FormDescription>
                Set to true to return the full state name with a capitalized
                first letter. Returns the 2-letter state code by default.
              </FormDescription>
            </div>
            <div className="flex flex-col ">
              <div className="justify-end flex">
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                    disabled={isDisabled}
                  />
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
