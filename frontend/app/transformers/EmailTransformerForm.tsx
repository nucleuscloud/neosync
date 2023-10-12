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
import { ReactElement, useEffect, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function EmailTransformerForm(props: Props): ReactElement {
  const { index, setIsSheetOpen } = props;

  const fc = useFormContext();

  const vals = fc.getValues();

  //sheet re-renders on every open which resets state, so have to get the values from the mappings so user values persist across sheet openings
  const [pd, setPd] = useState<boolean>(
    vals.mappings[index ?? 0].transformer.config.preserveDomain
  );
  const [pl, setPl] = useState<boolean>(
    vals.mappings[index ?? 0].transformer.config.preserveLength
  );

  const handleSubmit = () => {
    fc.setValue(`mappings.${index}.transformer.config.preserveDomain`, pd, {
      shouldValidate: false,
    });
    fc.setValue(`mappings.${index}.transformer.config.preserveLength`, pl, {
      shouldValidate: false,
    });
    setIsSheetOpen!(false);
  };

  //since component is in a controlled state, have to manually handle closing the sheet when the user presses escape
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setIsSheetOpen!(false);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    // Clean up the event listener when the component unmounts
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, []);

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.preserveLength`}
        defaultValue={pl}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Preserve Length</FormLabel>
              <FormDescription>
                Set the length of the output email to be the same as the input
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={pl}
                onCheckedChange={() => {
                  pl ? setPl(false) : setPl(true);
                }}
              />
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.preserveDomain`}
        defaultValue={pd}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Preserve Domain</FormLabel>
              <FormDescription>
                Set the length of the output email to be the same as the input
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={pd}
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
