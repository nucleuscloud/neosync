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

export default function TransformFullNameForm(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();

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
                Generates a full name which has the same first name and last
                name length as the input first and last names
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={pl}
                disabled={isUserDefinedTransformer(transformer)}
                onCheckedChange={() => {
                  pl ? setPl(false) : setPl(true);
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
