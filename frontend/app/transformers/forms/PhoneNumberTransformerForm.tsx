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
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function PhoneNumberTransformerForm(props: Props): ReactElement {
  const { index, setIsSheetOpen } = props;

  const fc = useFormContext();

  const vals = fc.getValues();

  //sheet re-renders on every open which resets state, so have to get the values from the mappings so user values persist across sheet openings
  const [ih, setIh] = useState<boolean>(
    vals.mappings[index ?? 0].transformer.config.includeHyphens
  );
  const [e164, setE164] = useState<boolean>(
    vals.mappings[index ?? 0].transformer.config.e164Format
  );
  const [pl, setPl] = useState<boolean>(
    vals.mappings[index ?? 0].transformer.config.preserveLength
  );

  const handleSubmit = () => {
    fc.setValue(`mappings.${index}.transformer.config.preserveLength`, pl, {
      shouldValidate: false,
    });
    fc.setValue(`mappings.${index}.transformer.config.e164Format`, e164, {
      shouldValidate: false,
    });
    fc.setValue(`mappings.${index}.transformer.config.includeHyphens`, ih, {
      shouldValidate: false,
    });
    setIsSheetOpen!(false);
  };

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
                Set the length of the output phone number to be the same as the
                input
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
        name={`mappings.${index}.transformer.config.includeHyphens`}
        defaultValue={ih}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Include Hyphens</FormLabel>
              <FormDescription>
                Include hyphens in the output phone number. Note: this only
                works with 10 digit phone numbers.
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={ih}
                onCheckedChange={() => {
                  ih ? setIh(false) : setIh(true);
                }}
              />
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.e164Format`}
        defaultValue={e164}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Format in E164 Format</FormLabel>
              <FormDescription>
                Format the output phone number in the E164 Format. For ex.
                +1892393573894
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={e164}
                onCheckedChange={() => {
                  e164 ? setE164(false) : setE164(true);
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
