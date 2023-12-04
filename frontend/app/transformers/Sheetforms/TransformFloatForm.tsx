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
import { UserDefinedTransformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';

interface Props {
  index?: number;
  setIsSheetOpen?: (val: boolean) => void;
  transformer: UserDefinedTransformer;
}

export default function TransformFloatForm(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();

  const plValue = fc.getValues(
    `mappings.${index}.transformer.config.config.value.preserveLength`
  );

  const [pl, setPl] = useState<boolean>(plValue);

  const psValue = fc.getValues(
    `mappings.${index}.transformer.config.config.value.preserveSign`
  );

  const [ps, setPs] = useState<boolean>(psValue);

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.preserveLength`,
      pl,
      {
        shouldValidate: false,
      }
    );
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.preserveSign`,
      ps,
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
                Set the length of the output first name to be the same as the
                input
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={pl}
                disabled={transformer.id ? true : false}
                onCheckedChange={() => {
                  pl ? setPl(false) : setPl(true);
                }}
              />
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.config.value.preserveSign`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Preserve Sign</FormLabel>
              <FormDescription className="w-[90%]">
                Preserve the sign of the input float to the output float. For
                example, if the input float is positive then the output float
                will also be positive.
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={ps}
                disabled={transformer.id ? true : false}
                onCheckedChange={() => {
                  ps ? setPs(false) : setPs(true);
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
