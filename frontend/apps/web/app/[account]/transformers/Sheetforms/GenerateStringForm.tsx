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
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function GenerateStringForm(props: Props): ReactElement {
  const { index, setIsSheetOpen } = props;

  const fc = useFormContext();

  const minValue = fc.getValues(
    `mappings.${index}.transformer.config.config.value.min`
  );
  const [min, setMin] = useState<number>(minValue);

  const maxVal = fc.getValues(
    `mappings.${index}.transformer.config.config.value.max`
  );
  const [max, setMax] = useState<number>(maxVal);
  const [disableSave, setDisableSave] = useState<boolean>(false);
  const [minError, setMinError] = useState<string>('');
  const [maxError, setMaxError] = useState<string>('');

  const handleSubmit = () => {
    fc.setValue(`mappings.${index}.transformer.config.config.value.min`, min, {
      shouldValidate: false,
    });
    fc.setValue(`mappings.${index}.transformer.config.config.value.max`, max, {
      shouldValidate: false,
    });
    setIsSheetOpen!(false);
  };

  const handleSettingMinRange = (value: number) => {
    if (value < 1 || value > max) {
      setMinError(
        'Minimum length cannot be less than 1 or greater than the max length'
      );
      setMin(value);
      setDisableSave(true);
    } else {
      setMinError('');
      setDisableSave(false);
      setMin(value);
    }
  };
  const handleSettingMaxRange = (value: number) => {
    if (value < min) {
      setMaxError(
        'Maximum length cannot be greater than 15 or less than the min length'
      );
      setMax(value);
      setDisableSave(true);
    } else {
      setMaxError('');
      setDisableSave(false);
      setMax(value);
    }
  };

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.config.value.min`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Minimum Length</FormLabel>
              <FormDescription>
                Set the minimum length range of the output string.
              </FormDescription>
            </div>
            <FormControl>
              <div className="max-w-[180px]">
                <Input
                  type="number"
                  className="max-w-[180px]"
                  value={String(min)}
                  onChange={(event) =>
                    handleSettingMinRange(Number(event.target.value))
                  }
                />
                <FormError errorMessage={minError} />
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
              <FormDescription>
                Set the maximum length range of the output string.
              </FormDescription>
            </div>
            <FormControl>
              <div className="max-w-[180px]">
                <Input
                  value={String(max)}
                  onChange={(event) =>
                    handleSettingMaxRange(Number(event.target.value))
                  }
                />
                <FormError errorMessage={maxError} />
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
