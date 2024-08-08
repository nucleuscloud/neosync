'use client';
import { FormDescription, FormLabel } from '@/components/ui/form';

import FormErrorMessage from '@/components/FormErrorMessage';
import { Switch } from '@/components/ui/switch';
import { PlainMessage } from '@bufbuild/protobuf';
import { GenerateCardNumber } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props
  extends TransformerConfigProps<
    GenerateCardNumber,
    PlainMessage<GenerateCardNumber>
  > {}

export default function GenerateCardNumberForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

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
      <FormErrorMessage message={errors?.validLuhn?.message} />
    </div>
  );
}
