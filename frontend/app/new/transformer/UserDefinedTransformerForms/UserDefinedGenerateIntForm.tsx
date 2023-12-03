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
import { ReactElement } from 'react';
import { useFormContext } from 'react-hook-form';

interface Props {
  isDisabled?: boolean;
}

export default function CustomGenerateIntForm(props: Props): ReactElement {
  const fc = useFormContext();

  const digitLength = Array.from({ length: 18 }, (_, index) => index + 1);

  const signs = ['positive', 'negative', 'random'];

  const { isDisabled } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`config.config.value.sign`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Sign</FormLabel>
              <FormDescription>
                Set the sign of the output integer.
              </FormDescription>
            </div>
            <FormControl>
              <Select
                onValueChange={field.onChange}
                value={String(field.value)}
                disabled={isDisabled}
              >
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="positive" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {signs.map((item) => (
                      <SelectItem value={item} key={item}>
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
        name={`config.config.value.length`}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Length</FormLabel>
              <FormDescription>
                Set the length of the output integer. The default length is 4.
                The max length is 18.
              </FormDescription>
            </div>
            <FormControl>
              <Select
                disabled={isDisabled}
                onValueChange={field.onChange}
                value={String(field.value)}
              >
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="4" />
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
