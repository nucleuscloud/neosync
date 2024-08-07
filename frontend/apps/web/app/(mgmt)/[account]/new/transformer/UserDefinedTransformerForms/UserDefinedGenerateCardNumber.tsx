'use client';
import { FormDescription, FormLabel } from '@/components/ui/form';

import { Switch } from '@/components/ui/switch';
import { GenerateCardNumber } from '@neosync/sdk';
import { ReactElement } from 'react';

interface Props {
  value: GenerateCardNumber;
  setValue(value: GenerateCardNumber): void;
  isDisabled?: boolean;
}

export default function UserDefinedGenerateCardNumberForm(
  props: Props
): ReactElement {
  const { value, setValue, isDisabled } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <div className="space-y-0.5">
        <FormLabel>Valid Luhn</FormLabel>
        <FormDescription className="w-[90%]">
          Generate a 16 digit card number that passes a luhn check.
        </FormDescription>
      </div>
      <Switch
        checked={value.validLuhn}
        onCheckedChange={(checked) => {
          setValue(new GenerateCardNumber({ ...value, validLuhn: checked }));
        }}
        disabled={isDisabled}
      />
    </div>
  );
}
