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
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';

import { Switch } from '@/components/ui/switch';
import { RandomString_StringCase } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
interface Props {
  index?: number;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function RandomStringTransformerForm(
  props: Props
): ReactElement {
  const { index, setIsSheetOpen } = props;

  const fc = useFormContext();

  const vals = fc.getValues();

  //sheet re-renders on every open which resets state, so have to get the values from the mappings so user values persist across sheet openings
  const [pl, setPl] = useState<boolean>(
    vals.mappings[index ?? 0].transformer.config.preserveLength
  );
  const [sc, setSc] = useState<RandomString_StringCase>(
    vals.mappings[index ?? 0].transformer.config.strCase
  );

  const [sl, setSl] = useState<number>(
    vals.mappings[index ?? 0].transformer.config.strLength ?? 0
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
                className="max-w-[180px]"
                placeholder="10"
                disabled={pl}
                onChange={(event) => {
                  setSl(Number(event.target.value));
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
              <RadioGroup
                defaultValue="option-one"
                className="flex flex-row"
                onValueChange={(value: string) => setSc(StrCaseToType(value))}
              >
                <div className="flex flex-row items-center space-x-2">
                  <RadioGroupItem value="lower" id="lower" />
                  <Label htmlFor="lower">lower</Label>
                </div>
                <div className="flex items-center space-x-2">
                  <RadioGroupItem value="upper" id="upper" />
                  <Label htmlFor="upper">UPPER</Label>
                </div>
                <div className="flex items-center space-x-2">
                  <RadioGroupItem value="title" id="title" />
                  <Label htmlFor="title">Title</Label>
                </div>
              </RadioGroup>
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

function StrCaseToType(s: string): RandomString_StringCase {
  switch (s) {
    case 'lower':
      return RandomString_StringCase.LOWER;
    case 'upper':
      return RandomString_StringCase.UPPER;
    case 'title':
      return RandomString_StringCase.TITLE;
  }

  return RandomString_StringCase.UPPER;
}
