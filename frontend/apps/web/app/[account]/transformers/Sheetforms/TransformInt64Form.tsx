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
import { Transformer, isUserDefinedTransformer } from '@/shared/transformers';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';

interface Props {
  index?: number;
  setIsSheetOpen?: (val: boolean) => void;
  transformer: Transformer;
}

export default function TransformInt64Form(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;
  const fc = useFormContext();

  const rMinVal = fc.getValues(
    `mappings.${index}.transformer.config.config.value.randomizationRangeMin`
  );

  const [rMin, setRMin] = useState<number>(rMinVal);

  const rMaxVal = fc.getValues(
    `mappings.${index}.transformer.config.config.value.randomizationRangeMax`
  );

  const [rMax, setRMax] = useState<number>(rMaxVal);
  const [disableSave, setDisableSave] = useState<boolean>(false);
  const [rMinError, setRMinError] = useState<string>('');
  const [rMaxError, setRMaxError] = useState<string>('');

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.randomizationRangeMin`,
      rMin,
      {
        shouldValidate: false,
      }
    );
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.randomizationRangeMax`,
      rMax,
      {
        shouldValidate: false,
      }
    );
    setIsSheetOpen!(false);
  };

  const handleSettingMinRange = (value: number) => {
    if (value > rMax) {
      setRMinError('Minimum bound cannot be greater than the max bound');
      setRMin(value);
      setDisableSave(true);
    } else {
      setRMinError('');
      setDisableSave(false);
      setRMin(value);
    }
  };
  const handleSettingMaxRange = (value: number) => {
    if (value < rMin) {
      setRMaxError('Maximum bound cannot be less than the min bound');
      setRMax(value);
      setDisableSave(true);
    } else {
      setRMaxError('');
      setDisableSave(false);
      setRMax(value);
    }
  };

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.config.value.randomizationRangeMin`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Minimum Value</FormLabel>
              <FormDescription className="w-[90%]">
                Sets a minium lower range value. This will create an lowerbound
                around the source input value. For example, if the input value
                is 10, and you set this value to 5, then the maximum range will
                be 5.
              </FormDescription>
            </div>
            <FormControl>
              <div className="max-w-[180px]">
                <Input
                  value={String(rMin)}
                  disabled={isUserDefinedTransformer(transformer)}
                  onChange={(event) =>
                    handleSettingMinRange(Number(event.target.value))
                  }
                />
                <FormError errorMessage={rMinError} />
              </div>
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.config.value.randomizationRangeMax`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Maxiumum Range Value</FormLabel>
              <FormDescription className="w-[90%]">
                Sets a maximum upper range value. This will create an upperbound
                around the source input value. For example, if the input value
                is 10, and you set this value to 5, then the maximum range will
                be 15.
              </FormDescription>
            </div>
            <FormControl>
              <div className="max-w-[180px]">
                <Input
                  value={String(rMax)}
                  disabled={isUserDefinedTransformer(transformer)}
                  onChange={(event) =>
                    handleSettingMaxRange(Number(event.target.value))
                  }
                />
                <FormError errorMessage={rMaxError} />
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
