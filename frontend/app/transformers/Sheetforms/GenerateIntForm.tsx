'use client';
import FormError from '@/components/FormError';
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
import { ReactElement, useEffect, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function GenerateIntForm(props: Props): ReactElement {
  const { index, setIsSheetOpen } = props;

  const fc = useFormContext();

  const signs = ['positive', 'negative', 'random'];

  const sValue = fc.getValues(
    `mappings.${index}.transformer.config.config.value.sign`
  );

  const [sign, setSign] = useState<string>(sValue);

  const lValue = fc.getValues(
    `mappings.${index}.transformer.config.config.value.length`
  );

  const [length, setLength] = useState<number>(lValue);
  const [disableSave, setDisableSave] = useState<boolean>(false);
  const [lengthError, setLengthError] = useState<string>('');

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.sign`,
      sign,
      {
        shouldValidate: false,
      }
    );
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.length`,
      length,
      {
        shouldValidate: false,
      }
    );
    setIsSheetOpen!(false);
  };

  useEffect(() => {
    if (length > 18 || length < 1) {
      setDisableSave(true);
      setLengthError('1 < length < 18.');
    } else {
      setDisableSave(false);
      setLengthError('');
    }
  }, [setLength, length]);

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.config.value.sign`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Sign</FormLabel>
              <FormDescription className="w-[90%]">
                Set the sign of the generated integer value. You can select
                Random in order to randomize the sign.
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
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.config.value.length`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Length</FormLabel>
              <FormDescription className="w-[90%]">
                Set the length of the output integer. The default length is 4.
                The max length is 18 digits.
              </FormDescription>
            </div>
            <FormControl>
              <div className="max-w-[180px]">
                <Input
                  type="number"
                  placeholder="4"
                  max={18}
                  value={String(length)}
                  onChange={(e) => setLength(Number(e.target.value))}
                />
                <FormError errorMessage={lengthError} />
              </div>
            </FormControl>
          </FormItem>
        )}
      />
      <div className="flex justify-end">
        <Button type="button" onClick={handleSubmit} disabled={disableSave}>
          Save
        </Button>
      </div>
    </div>
  );
}
