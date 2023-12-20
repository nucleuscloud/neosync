'use client';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';
import { Switch } from '@/components/ui/switch';
import { ReactElement } from 'react';
import { useFormContext } from 'react-hook-form';
import {
  CreateUserDefinedTransformerSchema,
  UpdateUserDefinedTransformer,
} from '../schema';

interface Props {
  isDisabled?: boolean;
}

export default function UserDefinedGenerateUuidForm(
  props: Props
): ReactElement {
  const fc = useFormContext<
    UpdateUserDefinedTransformer | CreateUserDefinedTransformerSchema
  >();

  const { isDisabled } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`config.value.includeHyphens`}
        control={fc.control}
        disabled={isDisabled}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Include hyphens</FormLabel>
              <FormDescription className="w-[90%]">
                Set to true to include hyphens in the generated UUID. Note: some
                databases such as Postgres automatically convert UUIDs with no
                hyphens to have hyphens when they store the data.
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={field.value}
                onCheckedChange={field.onChange}
                disabled={isDisabled}
              />
            </FormControl>
          </FormItem>
        )}
      />
    </div>
  );
}
