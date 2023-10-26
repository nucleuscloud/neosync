'use client';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';
import { Switch } from '@/components/ui/switch';
import {
  FullName,
  Transformer,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement } from 'react';
import { useFormContext } from 'react-hook-form';

interface Props {
  transformer: Transformer;
}
export default function CustomFullNameTransformerForm(
  props: Props
): ReactElement {
  const { transformer } = props;

  const fc = useFormContext();

  const t = transformer.config?.config.value as FullName;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`transformerConfig.preserveLength`}
        defaultValue={t.preserveLength}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Preserve Length</FormLabel>
              <FormDescription>
                Generates a full name which has the same first name and last
                name length as the input first and last names
              </FormDescription>
            </div>
            <FormControl>
              <Switch checked={field.value} onCheckedChange={field.onChange} />
            </FormControl>
          </FormItem>
        )}
      />
    </div>
  );
}
