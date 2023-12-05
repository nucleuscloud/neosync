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
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function GenerateStringForm(props: Props): ReactElement {
  const { index, setIsSheetOpen } = props;

  const fc = useFormContext();

  const sLength = fc.getValues(
    `mappings.${index}.transformer.config.config.value.length`
  );

  const [sl, setSl] = useState<number>(sLength);

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.length`,
      sl,
      {
        shouldValidate: false,
      }
    );
    setIsSheetOpen!(false);
  };

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.config.value.length`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Length</FormLabel>
              <FormDescription className="w-[90%]">
                Set the length of the output string. The default length is 6.
              </FormDescription>
            </div>
            <FormControl>
              <Input
                type="number"
                className="max-w-[180px]"
                placeholder="6"
                value={String(sl)}
                onChange={(event) => {
                  setSl(Number(event.target.value));
                }}
              />
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
