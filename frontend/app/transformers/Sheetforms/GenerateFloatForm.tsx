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
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import {
  GenerateFloat,
  UserDefinedTransformer,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement, useState } from 'react';
import { Controller, useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  transformer: UserDefinedTransformer;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function GenerateFloatForm(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();

  const config = transformer?.config?.config.value as GenerateFloat;

  console.log('transformer config', config);

  const s = fc.getValues(
    `mappings.${index}.transformer.config.config.value.sign`
  );

  const [sign, setSign] = useState<string>(s ? s : config?.sign);

  const bdValue = fc.getValues(
    `mappings.${index}.transformer.config.config.value.digitsBeforeDecimal`
  );
  const [bd, setBd] = useState<number>(
    bdValue ? bdValue : Number(config.digitsBeforeDecimal)
  );

  const adValue = fc.getValues(
    `mappings.${index}.transformer.config.config.value.digitsAfterDecimal`
  );
  const [ad, setAd] = useState<number>(
    adValue ? adValue : Number(config.digitsAfterDecimal)
  );

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.sign`,
      sign,
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

  const signs = ['positive', 'negative', 'random'];

  console.log('sign', sign);

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <Controller
        name={`mappings.${index}.transformer.config.config.value.sign`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0 z-10">
              <FormLabel>Sign</FormLabel>
              <FormDescription className="w-[90%]">
                Set the sign of the generated float value. You can select Random
                in order to randomize the sign.
              </FormDescription>
            </div>
            <FormControl>
              <RadioGroup
                onValueChange={(val: string) => {
                  setSign(val);
                }}
                defaultValue={sign}
                value={sign}
                className="flex flex-col space-y-1 justify-left w-[25%]"
              >
                {signs.map((item) => (
                  <FormItem
                    className="flex items-center space-x-3 space-y-0"
                    key={item}
                  >
                    <FormControl>
                      <RadioGroupItem value={item} />
                    </FormControl>
                    <FormLabel className="font-normal">{item}</FormLabel>
                  </FormItem>
                ))}
              </RadioGroup>
              {/* <Select
                onValueChange={(val: string) => {
                  setSign(val);
                }}
                value={sign}
              >
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="3" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {signs.map((item) => (
                      <SelectItem value={item} key={item}>
                        {item}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select> */}
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
              <FormDescription className="w-[90%]">
                Set the number of digits you want the float to have before the
                decimal place. For example, a value of 6 will result in a float
                that is 6 digits long. This has a max of 8 digits.
              </FormDescription>
            </div>
            <FormControl>
              <Input
                className="max-w-[180px]"
                placeholder="3"
                max={9}
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
              <FormDescription className="w-[90%]">
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
