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
import {
  CustomTransformer,
  RandomFloat,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  transformer: CustomTransformer;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function RandomFloatTransformerForm(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();

  const config = transformer?.config?.config.value as RandomFloat;

  const [pl, setPl] = useState<boolean>(
    config?.preserveLength ? config?.preserveLength : false
  );

  const [bd, setBd] = useState<number>(
    config?.digitsBeforeDecimal ? Number(config.digitsBeforeDecimal) : 0
  );

  const [ad, setAd] = useState<number>(
    config?.digitsAfterDecimal ? Number(config.digitsAfterDecimal) : 0
  );

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.preserveLength`,
      pl,
      {
        shouldValidate: false,
      }
    );
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.digitsBeforeDecimal`,
      bd,
      {
        shouldValidate: false,
      }
    );
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.digitsAfterDecimal`,
      ad,
      {
        shouldValidate: false,
      }
    );
    setIsSheetOpen!(false);
  };

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.config.value.preserveLength`}
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
        name={`mappings.${index}.transformer.config.config.value.digitsBeforeDecimal`}
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
                value={bd}
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
        name={`mappings.${index}.transformer.config.config.value.digitsAfterDecimal`}
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
                value={ad}
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
