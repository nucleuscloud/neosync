'use client';
import { Button } from '@/components/ui/button';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { GenerateCategorical } from '@neosync/sdk';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function GenerateCategoricalForm(props: Props): ReactElement {
  const { index, setIsSheetOpen } = props;

  const fc = useFormContext();

  const catValue = fc.getValues(
    `mappings.${index}.transformer.config.value.categories`
  );

  const [categories, setCategories] = useState<string>(catValue);

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.value`,
      new GenerateCategorical({ categories: categories }),
      {
        shouldValidate: false,
      }
    );
    setIsSheetOpen!(false);
  };

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.value.categories`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Categories</FormLabel>
              <FormDescription>
                Provide a list of comma-separated string values that you want to
                randomly select from.
              </FormDescription>
            </div>
            <FormControl>
              <div className="max-w-[180px]">
                <Input
                  type="string"
                  className="max-w-[180px]"
                  value={categories}
                  onChange={(event) => setCategories(event.target.value)}
                />
              </div>
            </FormControl>
          </FormItem>
        )}
      />
      <div className="flex justify-end">
        <Button type="button" onClick={handleSubmit}>
          Save
        </Button>
      </div>
    </div>
  );
}
