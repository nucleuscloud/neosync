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
import { UserDefinedTransformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement, useEffect, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  transformer: UserDefinedTransformer;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function GenerateE164NumberForm(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();

  const lValue = fc.getValues(
    `mappings.${index}.transformer.config.config.value.length`
  );
  const [length, setLength] = useState<number>(lValue);
  const [disableSave, setDisableSave] = useState<boolean>(false);
  const [lengthError, setLengthError] = useState<string>('');

  const handleSubmit = () => {
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
    if (length > 15 || length < 9) {
      setDisableSave(true);
      setLengthError('9 < length < 15.');
    } else {
      setDisableSave(false);
      setLengthError('');
    }
  }, [setLength, length]);

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.config.value.length`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Length</FormLabel>
              <FormDescription className="w-[90%]">
                Set the length of the output phone number to be the same as the
                input
              </FormDescription>
            </div>
            <FormControl>
              <div className="max-w-[180px]">
                <Input
                  placeholder="3"
                  max={9}
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
