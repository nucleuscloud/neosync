'use client';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { ReactElement } from 'react';
import { useFormContext } from 'react-hook-form';
import {
  CreateUserDefinedTransformerSchema,
  UpdateUserDefinedTransformer,
} from '../schema';
interface Props {
  isDisabled?: boolean;
}

export default function UserDefinedGenerateCategoricalForm(
  props: Props
): ReactElement {
  const fc = useFormContext<
    UpdateUserDefinedTransformer | CreateUserDefinedTransformerSchema
  >();

  const { isDisabled } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`config.value.categories`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Categories</FormLabel>
              <FormDescription>
                Provide a list of comma-separated string values that you want to
                randomly select from.
              </FormDescription>
            </div>
            <FormControl>
              <div className="w-[600px]">
                <Input
                  value={field.value}
                  type="string"
                  onChange={field.onChange}
                  disabled={isDisabled}
                />
              </div>
            </FormControl>
          </FormItem>
        )}
      />
    </div>
  );
}
