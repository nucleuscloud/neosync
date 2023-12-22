'use client';
import { Button } from '@/components/ui/button';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
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

  const precisionVal = fc.getValues(
    `mappings.${index}.transformer.config.value.precision`
  );
  const [precision, setPrecision] = useState<number>(precisionVal);

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.value`,
      new GenerateFloat64({
        randomizeSign: sign,
        min,
        max,
        precision: BigInt(precision),
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
            <FormMessage />
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
            <FormMessage />
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
            <FormMessage />
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.value.precision`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Precision</FormLabel>
              <FormDescription>
                Sets the precision for the entire float64 value, not just the
                decimals. For example. a precision of 4 would update a float64
                value of 23.567 to 23.56.
              </FormDescription>
            </div>

            <FormControl>
              <Input
                type="number"
                className="max-w-[180px]"
                value={String(precision)}
                onChange={(event) => {
                  setPrecision(Number(event.target.value));
                }}
              />
            </FormControl>
            <FormMessage />
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
