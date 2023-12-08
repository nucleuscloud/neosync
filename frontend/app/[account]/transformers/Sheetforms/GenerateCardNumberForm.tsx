'use client';
import { Button } from '@/components/ui/button';
import {
  FormControl,
  FormDescription,
  FormItem,
  FormLabel,
} from '@/components/ui/form';

import { Switch } from '@/components/ui/switch';
import { Transformer, isUserDefinedTransformer } from '@/shared/transformers';
import { ReactElement, useState } from 'react';
import { Controller, useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  transformer: Transformer;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function GenerateCardNumberForm(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();

  const vlValue = fc.getValues(
    `mappings.${index}.transformer.config.config.value.validLuhn`
  );
  const [vl, setVl] = useState<boolean>(vlValue);

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.validLuhn`,
      vl,
      {
        shouldValidate: false,
      }
    );
    setIsSheetOpen!(false);
  };

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <Controller
        name={`mappings.${index}.transformer.config.config.value.validLuhn`}
        defaultValue={vl}
        disabled={isUserDefinedTransformer(transformer)}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Valid Luhn</FormLabel>
              <FormDescription className="w-[90%]">
                Generate a 16 digit card number that passes a luhn check.
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={vl}
                onCheckedChange={() => {
                  vl ? setVl(false) : setVl(true);
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
