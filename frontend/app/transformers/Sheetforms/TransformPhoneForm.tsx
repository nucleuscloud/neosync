'use client';
import { Button } from '@/components/ui/button';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';
import { Switch } from '@/components/ui/switch';
import { Transformer, isUserDefinedTransformer } from '@/shared/transformers';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  transformer: Transformer;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function TransformPhoneForm(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();
  const ihValue = fc.getValues(
    `mappings.${index}.transformer.config.config.value.includeHyphens`
  );

  const [ih, setIh] = useState<boolean>(ihValue);
  const plValue = fc.getValues(
    `mappings.${index}.transformer.config.config.value.preserveLength`
  );

  const [pl, setPl] = useState<boolean>(plValue);

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.preserveLength`,
      pl,
      {
        shouldValidate: false,
      }
    );
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.includeHyphens`,
      ih,
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
              <FormDescription className="w-[90%]">
                Set the length of the output phone number to be the same as the
                input
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={pl}
                disabled={ih || isUserDefinedTransformer(transformer)}
                onCheckedChange={() => setPl(!pl)}
              />
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.config.value.includeHyphens`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Include Hyphens</FormLabel>
              <FormDescription className="w-[90%]">
                Include hyphens in the output phone number. Note: this only
                works with 10 digit phone numbers.
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={ih}
                disabled={pl || isUserDefinedTransformer(transformer)}
                onCheckedChange={() => setIh(!ih)}
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
