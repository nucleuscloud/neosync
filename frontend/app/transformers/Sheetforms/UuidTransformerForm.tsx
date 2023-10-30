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
  Uuid,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  transformer: CustomTransformer;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function UuidTransformerForm(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();

  const config = transformer?.config?.config.value as Uuid;

  const [ih, setIh] = useState<boolean>(
    config?.includeHyphen ? config?.includeHyphen : false
  );

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.config.value.includeHyphen`,
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
        name={`mappings.${index}.transformer.config.config.value.includeHyphen`}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Include hyphens</FormLabel>
              <FormDescription>
                Set to true to include hyphens in the generated UUID. Note: some
                databases such as Postgres automatically convert UUIDs with no
                hyphens to have hypthens when they store the data.
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
