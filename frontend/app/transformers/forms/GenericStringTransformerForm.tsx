'use client';
import { Button } from '@/components/ui/button';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { GenericString_StringCase } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function GenericStringTransformerForm(
  props: Props
): ReactElement {
  const { index, setIsSheetOpen } = props;

  const fc = useFormContext();

  const vals = fc.getValues();

  //sheet re-renders on every open which resets state, so have to get the values from the mappings so user values persist across sheet openings
  const [pl, setPl] = useState<boolean>(
    vals.mappings[index ?? 0].transformer.config.preserveLength
  );
  const [sc, setSc] = useState<GenericString_StringCase>(
    vals.mappings[index ?? 0].transformer.config.strCase
  );

  const [sl, setSl] = useState<number>(
    vals.mappings[index ?? 0].transformer.config.strLength
  );

  const handleSubmit = () => {
    fc.setValue(`mappings.${index}.transformer.config.strCase`, sc, {
      shouldValidate: false,
    });
    fc.setValue(`mappings.${index}.transformer.config.preserveLength`, pl, {
      shouldValidate: false,
    });
    fc.setValue(`mappings.${index}.transformer.config.strLength`, sl, {
      shouldValidate: false,
    });
    setIsSheetOpen!(false);
  };

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.preserveLength`}
        defaultValue={pl}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Preserve Length</FormLabel>
              <FormDescription>
                Set the length of the output string to be the same as the input
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
        name={`mappings.${index}.transformer.config.strLength`}
        defaultValue={sl}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>String Length</FormLabel>
              <FormDescription>
                Set the length of the output string. The default length is 10.
              </FormDescription>
            </div>
            <FormControl>
              <Input
                type="number"
                value={sl}
                className="max-w-[180px]"
                disabled={pl}
                placeholder="10"
                onChange={() => {
                  sl ? setSl(10) : setSl(sl);
                }}
              />
            </FormControl>
          </FormItem>
        )}
      />
      <FormField
        name={`mappings.${index}.transformer.config.strCase`}
        defaultValue={sc}
        render={() => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>String Case</FormLabel>
              <FormDescription>
                Set the case for the output string.
              </FormDescription>
            </div>
            <FormControl>
              <Select
                onValueChange={(value: string) => {
                  setSc(StrCaseToType(value));
                }}
              >
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="lower" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="lower">lower</SelectItem>
                  <SelectItem value="upper">UPPER</SelectItem>
                  <SelectItem value="title">Title</SelectItem>
                </SelectContent>
              </Select>
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

function StrCaseToType(s: string): GenericString_StringCase {
  switch (s) {
    case 'lower':
      return GenericString_StringCase.LOWER;
    case 'upper':
      return GenericString_StringCase.UPPER;
    case 'title':
      return GenericString_StringCase.TITLE;
  }

  return GenericString_StringCase.UPPER;
}
