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
import { GenerateFloat64 } from '@neosync/sdk';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function GenerateFloat64Form(props: Props): ReactElement {
  const { index, setIsSheetOpen } = props;

  const fc = useFormContext();

  const s = fc.getValues(
    `mappings.${index}.transformer.config.value.randomizeSign`
  );
  const [sign, setSign] = useState<boolean>(s);

  const minValue = fc.getValues(
    `mappings.${index}.transformer.config.value.min`
  );
  const [min, setMin] = useState<number>(minValue);

  const maxVal = fc.getValues(`mappings.${index}.transformer.config.value.max`);
  const [max, setMax] = useState<number>(maxVal);

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.value`,
      new GenerateFloat64({
        randomizeSign: sign,
        min,
        max,
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
        name={`mappings.${index}.transformer.config.value.randomizeSign`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0 z-10">
              <FormLabel>Randomize Sign</FormLabel>
              <FormDescription>
                Randomly sets a sign to the generated float64 value. By default,
                it generates a positive number.
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={sign}
                onCheckedChange={() => {
                  setSign(!sign);
                }}
              />
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.value.min`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Minimum Value</FormLabel>
              <FormDescription>
                Sets a minimum range for generated float64 value. This can be
                negative as well.
              </FormDescription>
            </div>
            <FormControl>
              <Input
                className="max-w-[180px]"
                type="number"
                value={String(min)}
                onChange={(event) => {
                  setMin(Number(event.target.value));
                }}
              />
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.value.max`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Maxiumum Value</FormLabel>
              <FormDescription>
                Sets a maximum range for generated float64 value. This can be
                negative as well.
              </FormDescription>
            </div>
            <FormControl>
              <Input
                className="max-w-[180px]"
                type="number"
                value={String(max)}
                onChange={(event) => {
                  setMax(Number(event.target.value));
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
