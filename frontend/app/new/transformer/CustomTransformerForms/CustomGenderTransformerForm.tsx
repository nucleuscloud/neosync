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

export default function CustomGenderTransformerForm(): ReactElement {
  const fc = useFormContext();

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`config.config.value.abbreviate`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Abbreviate</FormLabel>
              <FormDescription>
                Abbreviate the gender to a single character. For example, female
                would be returned as f.
              </FormDescription>
            </div>
            <FormControl>
              <Switch checked={field.value} onCheckedChange={field.onChange} />
            </FormControl>
          </FormItem>
        )}
      />
    </div>
  );
}
