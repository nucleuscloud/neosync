'use client';
import SwitchCard from '@/components/switches/SwitchCard';
import {
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import { ReactElement } from 'react';

interface SubsetOptionsProps {
  maxColNum?: number;
}
export default function SubsetOptionsForm(
  props: SubsetOptionsProps
): ReactElement {
  const { maxColNum } = props;
  const grid = maxColNum ? `lg:grid-cols-${maxColNum}` : 'lg:grid-cols-3';
  return (
    <div className={`grid grid-cols-1 md:grid-cols-1 ${grid} gap-4`}>
      <div>
        <FormField
          name="subsetOptions.subsetByForeignKeyConstraints"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <SwitchCard
                  isChecked={field.value || false}
                  onCheckedChange={field.onChange}
                  title="Subset using foreign key constraints"
                  description="Subsets tables based on foreign key relationships"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>
    </div>
  );
}
