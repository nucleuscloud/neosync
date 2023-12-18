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
import { ReactElement, useEffect, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function GenerateE164PhoneNumberForm(
  props: Props
): ReactElement {
  const { index, setIsSheetOpen } = props;

  const fc = useFormContext();

  const minVal = fc.getValues(
    `mappings.${index}.transformer.config.config.value.min`
  );
  const maxVal = fc.getValues(
    `mappings.${index}.transformer.config.config.value.max`
  );
  const [min, setMin] = useState<number>(minVal);
  const [max, setMax] = useState<number>(maxVal);
  const [disableSave, setDisableSave] = useState<boolean>(false);
  const [lengthError, setLengthError] = useState<string>('');

  const handleSubmit = () => {
    fc.setValue(`mappings.${index}.transformer.config.config.value.min`, min, {
      shouldValidate: false,
    });
    setIsSheetOpen!(false);
  };

  fc.setValue(`mappings.${index}.transformer.config.config.value.max`, max, {
    shouldValidate: false,
  });
  setIsSheetOpen!(false);

  useEffect(() => {
    if (max > 15 || min < 9) {
      setDisableSave(true);
      setLengthError('9 < length < 15.');
    } else {
      setDisableSave(false);
      setLengthError('');
    }
  }, [min, max]);

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.config.value.min`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Minimum Length</FormLabel>
              <FormDescription className="w-[90%]">
                Set the minimum length of the output phone number. It cannot be
                less than 9.
              </FormDescription>
            </div>
            <FormControl>
              <div className="max-w-[180px]">
                <Input
                  placeholder="3"
                  max={9}
                  value={String(length)}
                  onChange={(e) => setMin(Number(e.target.value))}
                />
                <FormError errorMessage={lengthError} />
              </div>
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.config.value.max`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Maximum Length</FormLabel>
              <FormDescription className="w-[90%]">
                Set the maximum length of the output phone number. It cannot be
                greater than 15.
              </FormDescription>
            </div>
            <FormControl>
              <div className="max-w-[180px]">
                <Input
                  placeholder="3"
                  max={15}
                  value={String(length)}
                  onChange={(e) => setMax(Number(e.target.value))}
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
