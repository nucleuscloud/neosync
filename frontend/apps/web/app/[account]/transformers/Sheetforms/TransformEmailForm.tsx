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
import { TransformEmail } from '@neosync/sdk';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  transformer: Transformer;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function TransformEmailForm(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();

  const pdValue = fc.getValues(
    `mappings.${index}.transformer.config.value.preserveDomain`
  );

  const [pd, setPd] = useState<boolean>(pdValue);

  const plValue = fc.getValues(
    `mappings.${index}.transformer.config.value.preserveLength`
  );
  const [pl, setPl] = useState<boolean>(plValue);

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.value`,
      new TransformEmail({
        preserveDomain: pd,
        preserveLength: pl,
      }),
      {
        shouldValidate: false,
      }
    );
    setIsSheetOpen!(false);
  };

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.value.preserveLength`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Preserve Length</FormLabel>
              <FormDescription className="w-[90%]">
                Set the length of the output email to be the same as the input
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
      <FormField
        name={`mappings.${index}.transformer.config.value.preserveDomain`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Preserve Domain</FormLabel>
              <FormDescription className="w-[90%]">
                Preserve the input domain including top level domain to the
                output value. For ex. if the input is john@gmail.com, the output
                will be ij23o@gmail.com
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={pd}
                disabled={isUserDefinedTransformer(transformer)}
                onCheckedChange={() => {
                  pd ? setPd(false) : setPd(true);
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
