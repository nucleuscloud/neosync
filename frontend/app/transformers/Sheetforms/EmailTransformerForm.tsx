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
  EmailConfig,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  transformer: CustomTransformer;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function EmailTransformerForm(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();

  const config = transformer?.config?.config.value as EmailConfig;
  const [pd, setPd] = useState<boolean>(
    config?.preserveDomain ? config?.preserveDomain : false
  );
  const [pl, setPl] = useState<boolean>(
    config?.preserveLength ? config?.preserveLength : false
  );

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.preserveDomain`,
      pd,
      {
        shouldValidate: false,
      }
    );
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.preserveLength`,
      pl,
      {
        shouldValidate: false,
      }
    );

    fc.setValue(
      `mappings.${index}.transformer.config.config.case`,
      'emailConfig'
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
        name={`mappings.${index}.transformer.config.config.value.preserveDomain`}
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
