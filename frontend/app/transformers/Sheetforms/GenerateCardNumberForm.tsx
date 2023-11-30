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
import {
  CustomTransformer,
  GenerateCardNumber,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  transformer: CustomTransformer;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function GenerateCardNumberForm(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();

  const config = transformer?.config?.config.value as GenerateCardNumber;

  const [vl, setVl] = useState<boolean>(
    config?.validLuhn ? config?.validLuhn : false
  );

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
      <FormField
        name={`mappings.${index}.transformer.config.config.value.validLuhn`}
        defaultValue={vl}
        disabled={transformer.id ? true : false} //disable if a custom transformer; custom transformers have an ID field since they're stoerd in the db
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Valid Luhn</FormLabel>
              <FormDescription>
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
