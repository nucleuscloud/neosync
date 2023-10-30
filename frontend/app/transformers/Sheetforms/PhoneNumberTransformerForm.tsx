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
import {
  CustomTransformer,
  PhoneNumber,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  transformer: CustomTransformer;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function PhoneNumberTransformerForm(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();

  const vals = fc.getValues();

  const config = transformer?.config?.config.value as PhoneNumber;

  const [ih, setIh] = useState<boolean>(
    config?.includeHyphens ? config?.includeHyphens : false
  );
  const [e164, setE164] = useState<boolean>(
    config?.e164Format ? config?.e164Format : false
  );
  const [pl, setPl] = useState<boolean>(
    config?.preserveLength ? config?.preserveLength : false
  );

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.preserveLength`,
      pl,
      {
        shouldValidate: false,
      }
    );
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.e164Format`,
      e164,
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
              <FormDescription>
                Set the length of the output phone number to be the same as the
                input
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={pl}
                disabled={ih}
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
              <FormDescription>
                Include hyphens in the output phone number. Note: this only
                works with 10 digit phone numbers.
              </FormDescription>
            </div>
            <FormControl>
              <Switch
                checked={ih}
                disabled={e164 || pl}
                onCheckedChange={() => setIh(!ih)}
              />
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.config.value.e164Format`}
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
                disabled={ih}
                onCheckedChange={() => setE164(!e164)}
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
