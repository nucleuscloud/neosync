'use client';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';

import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import {
  RandomFloat,
  Transformer,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement } from 'react';
import { useFormContext } from 'react-hook-form';

interface Props {
  transformer: Transformer;
}

export default function CustomRandomFloatTransformerForm(
  props: Props
): ReactElement {
  const { transformer } = props;

  const fc = useFormContext();

  const t = transformer.config?.config.value as RandomFloat;

  const digitLength = Array.from({ length: 9 }, (_, index) => index + 1);

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
                Set the length of the output string to be the same as the input
              </FormDescription>
            </div>
            <FormControl>
              <Switch checked={field.value} onCheckedChange={field.onChange} />
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`transformerConfig.bd`}
        defaultValue={t.digitsBeforeDecimal}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Digits before decimal</FormLabel>
              <FormDescription className="w-[90%]">
                Set the number of digits you want the float to have before the
                decimal place. For example, a value of 6 will result in a float
                that is 6 digits long. This has a max of 8 digits.
              </FormDescription>
            </div>
            <FormControl>
              <Select
                disabled={fc.getValues('transformerConfig.preserveLength')}
                onValueChange={field.onChange}
                defaultValue={field.value}
              >
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder={String(t.digitsBeforeDecimal)} />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {digitLength.map((item) => (
                      <SelectItem value={String(item)} key={item}>
                        {item}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`transformerConfig.ad`}
        defaultValue={t.digitsAfterDecimal}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Digits after decimal</FormLabel>
              <FormDescription className="w-[90%]">
                Set the number of digits you want the float to have after the
                decimal place. For example, a value of 3 will result in a float
                that has 3 digits in the decimals place. This has a max of 8
                digits.
              </FormDescription>
            </div>
            <FormControl>
              <Select
                disabled={fc.getValues('transformerConfig.preserveLength')}
                onValueChange={field.onChange}
                defaultValue={field.value}
              >
                <SelectTrigger className="w-[190px]">
                  <SelectValue placeholder={String(t.digitsAfterDecimal)} />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {digitLength.map((item) => (
                      <SelectItem value={String(item)} key={item}>
                        {item}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </FormControl>
          </FormItem>
        )}
      />
    </div>
  );
}
