'use client';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';

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
  RandomString,
  RandomString_StringCase,
  Transformer,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { ReactElement } from 'react';
import { useFormContext } from 'react-hook-form';

interface Props {
  transformer: Transformer;
}

export default function CustomRandomStringTransformerForm(
  props: Props
): ReactElement {
  const { transformer } = props;

  const fc = useFormContext();

  const t = transformer.config?.config.value as RandomString;

  const digitLength = Array.from({ length: 12 }, (_, index) => index + 1);

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
        name={`transformerConfig.strLength`}
        defaultValue={t.strLength}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>String Length</FormLabel>
              <FormDescription>
                Set the length of the output string. The default length is 4.
                The max length is 18.
              </FormDescription>
            </div>
            <FormControl>
              <Select
                disabled={fc.getValues('transformerConfig.preserveLength')}
                onValueChange={field.onChange}
                defaultValue={field.value}
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
      <FormField
        name={`transformerConfig.strCase`}
        defaultValue={t.strCase}
        control={fc.control}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>String Case</FormLabel>
              <FormDescription>
                Set the case for the output string.
              </FormDescription>
            </div>
            <FormControl>
              <RadioGroup
                className="flex flex-row"
                onValueChange={field.onChange}
                defaultValue={field.value}
              >
                <div className="flex flex-row items-center space-x-2">
                  <RadioGroupItem
                    value={String(RandomString_StringCase.LOWER)}
                    id="lower"
                  />
                  <Label htmlFor="lower">lower</Label>
                </div>
                <div className="flex items-center space-x-2">
                  <RadioGroupItem
                    value={String(RandomString_StringCase.LOWER)}
                    id="upper"
                  />
                  <Label htmlFor="upper">UPPER</Label>
                </div>
                <div className="flex items-center space-x-2">
                  <RadioGroupItem
                    value={String(RandomString_StringCase.LOWER)}
                    id="title"
                  />
                  <Label htmlFor="title">Title</Label>
                </div>
              </RadioGroup>
            </FormControl>
          </FormItem>
        )}
      />
    </div>
  );
}
