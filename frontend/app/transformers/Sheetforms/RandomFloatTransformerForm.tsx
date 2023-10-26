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

import { Switch } from '@/components/ui/switch';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function RandomFloatTransformerForm(props: Props): ReactElement {
  const { index, setIsSheetOpen } = props;

  const fc = useFormContext();

  const vals = fc.getValues();

  //sheet re-renders on every open which resets state, so have to get the values from the mappings so user values persist across sheet openings
  const [pl, setPl] = useState<boolean>(
    vals.mappings[index ?? 0].transformer.config.preserveLength
  );

  const [bd, setBd] = useState<number>(
    vals.mappings[index ?? 0].transformer.config.bd ?? 0
  );

  const [ad, setAd] = useState<number>(
    vals.mappings[index ?? 0].transformer.config.ad ?? 0
  );

  const handleSubmit = () => {
    fc.setValue(`mappings.${index}.transformer.config.preserveLength`, pl, {
      shouldValidate: false,
    });
    fc.setValue(`mappings.${index}.transformer.config.bd`, bd, {
      shouldValidate: false,
    });
    fc.setValue(`mappings.${index}.transformer.config.ad`, ad, {
      shouldValidate: false,
    });
    setIsSheetOpen!(false);
  };

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.preserveLength`}
        defaultValue={pl}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Preserve Length</FormLabel>
              <FormDescription>
                Set the length of the output string to be the same as the input
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={pl}
                onCheckedChange={() => {
                  pl ? setPl(false) : setPl(true);
                }}
              />
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.bd`}
        defaultValue={bd}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Digits before decimal</FormLabel>
              <FormDescription>
                Set the number of digits you want the float to have before the
                decimal place. For example, a value of 6 will result in a float
                that is 6 digits long. This has a max of 8 digits.
              </FormDescription>
            </div>
            <FormControl>
              <Input
                className="max-w-[180px]"
                placeholder="10"
                max={9}
                disabled={pl}
                onChange={(event) => {
                  const inputValue = Math.min(
                    9,
                    Math.max(0, Number(event.target.value))
                  );
                  setBd(inputValue);
                }}
              />
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.ad`}
        defaultValue={ad}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Digits after decimal</FormLabel>
              <FormDescription>
                Set the number of digits you want the float to have after the
                decimal place. For example, a value of 3 will result in a float
                that has 3 digits in the decimals place. This has a max of 8
                digits.
              </FormDescription>
            </div>
            <FormControl>
              <Input
                className="max-w-[180px]"
                placeholder="10"
                disabled={pl}
                max={9}
                onChange={(event) => {
                  const inputValue = Math.min(
                    9,
                    Math.max(0, Number(event.target.value))
                  );
                  setAd(inputValue);
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
