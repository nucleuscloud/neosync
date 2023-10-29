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
  Transformer,
  Uuid,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement } from 'react';
import { useFormContext } from 'react-hook-form';

interface Props {
  transformer: Transformer;
}

export default function CustomUuidTransformerForm(props: Props): ReactElement {
  const { transformer } = props;

  const fc = useFormContext();

  const t = transformer.config?.config.value as Uuid;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`transformerConfig.includeHyphen`}
        defaultValue={t.includeHyphen}
        control={fc.control}
        render={({ field }) => (
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
              <Switch checked={field.value} onCheckedChange={field.onChange} />
            </FormControl>
          </FormItem>
        )}
      />
    </div>
  );
}
