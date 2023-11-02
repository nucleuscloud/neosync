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
import { ReactElement } from 'react';
import { useFormContext } from 'react-hook-form';

interface Props {
  isDisabled?: boolean;
}

export default function CustomRandomFloatTransformerForm(
  props: Props
): ReactElement {
  const fc = useFormContext();

  const digitLength = Array.from({ length: 9 }, (_, index) => index + 1);

  const { isDisabled } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`config.config.value.preserveLength`}
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
              <Switch
                checked={field.value}
                onCheckedChange={field.onChange}
                disabled={
                  fc.watch('config.config.value.digitsAfterDecimal') ||
                  fc.watch('config.config.value.digitsBeforeDecimal') ||
                  isDisabled
                }
              />
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`config.config.value.digitsBeforeDecimal`}
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
                disabled={
                  fc.watch('config.config.value.preserveLength') || isDisabled
                }
                onValueChange={field.onChange}
                value={field.value}
              >
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="2" />
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
        name={`config.config.value.digitsAfterDecimal`}
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
                disabled={
                  fc.watch('config.config.value.preserveLength') || isDisabled
                }
                onValueChange={field.onChange}
                value={field.value}
              >
                <SelectTrigger className="w-[190px]">
                  <SelectValue placeholder="3" />
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
