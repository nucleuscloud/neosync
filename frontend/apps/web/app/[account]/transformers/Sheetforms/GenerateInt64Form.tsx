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

export default function GenerateInt64Form(props: Props): ReactElement {
  const { index, setIsSheetOpen } = props;

  const fc = useFormContext();

  const s = fc.getValues(
    `mappings.${index}.transformer.config.config.value.randomizeSign`
  );
  const [sign, setSign] = useState<boolean>(s);

  const minValue = fc.getValues(
    `mappings.${index}.transformer.config.config.value.min`
  );
  const [min, setMin] = useState<number>(minValue);

  const maxVal = fc.getValues(
    `mappings.${index}.transformer.config.config.value.max`
  );
  const [max, setMax] = useState<number>(maxVal);

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.randomizeSign`,
      sign,
      {
        shouldValidate: false,
      }
    );
    fc.setValue(`mappings.${index}.transformer.config.config.value.min`, min, {
      shouldValidate: false,
    });

    fc.setValue(`mappings.${index}.transformer.config.config.value.max`, max, {
      shouldValidate: false,
    });
    setIsSheetOpen!(false);
  };

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.config.value.randomizeSign`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Randomize Sign</FormLabel>
              <FormDescription>
                Randomly the generated value between positive and negative. By
                default, it generates a positive number.
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={sign}
                onCheckedChange={() => {
                  sign ? setSign(false) : setSign(true);
                }}
              />
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.config.value.min`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Minimum Value</FormLabel>
              <FormDescription>
                Sets a minimum range for generated int64 value. This can be
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
        name={`mappings.${index}.transformer.config.config.value.max`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Maxiumum Value</FormLabel>
              <FormDescription>
                Sets a maximum range for generated int64 value. This can be
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
