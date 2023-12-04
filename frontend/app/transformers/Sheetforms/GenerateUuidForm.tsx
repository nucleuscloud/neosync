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
  transformer: UserDefinedTransformer;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function GenerateUuidForm(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();

  const ihValue = fc.getValues(
    `mappings.${index}.transformer.config.config.value.includeHyphens`
  );

  const [ih, setIh] = useState<boolean>(ihValue);

  const handleSubmit = () => {
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
        name={`mappings.${index}.transformer.config.config.value.includeHyphens`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Include hyphens</FormLabel>
              <FormDescription className="w-[90%]">
                Set to true to include hyphens in the generated UUID. Note: some
                databases such as Postgres automatically convert UUIDs with no
                hyphens to have hyphens when they store the data.
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
      <div className="flex justify-end">
        <Button type="button" onClick={handleSubmit}>
          Save
        </Button>
      </div>
    </div>
  );
}
