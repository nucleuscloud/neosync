'use client';
import { Button } from '@/components/ui/button';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';
import { Switch } from '@/components/ui/switch';
import { GenerateStringPhoneNumber } from '@neosync/sdk';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function GenerateStringPhoneNumberForm(
  props: Props
): ReactElement {
  const { index, setIsSheetOpen } = props;

  const fc = useFormContext();

  const ihValue = fc.getValues(
    `mappings.${index}.transformer.config.value.includeHyphens`
  );

  const [ih, setIh] = useState<boolean>(ihValue);
  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.value`,
      new GenerateStringPhoneNumber({
        includeHyphens: ih,
      }),
      {
        shouldValidate: false,
      }
    );
    setIsSheetOpen!(false);
  };

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.value.includeHyphens`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Include Hyphens</FormLabel>
              <FormDescription className="w-[90%]">
                Include hyphens in the output phone number. Note: this only
                works with 10 digit phone numbers.
              </FormDescription>
            </div>
            <FormControl>
              <Switch checked={ih} onCheckedChange={() => setIh(!ih)} />
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
