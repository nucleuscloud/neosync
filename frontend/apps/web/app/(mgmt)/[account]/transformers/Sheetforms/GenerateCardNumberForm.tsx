'use client';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';

import { Switch } from '@/components/ui/switch';
import { yupResolver } from '@hookform/resolvers/yup';
import { GenerateCardNumber } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { TRANSFORMER_SCHEMA_CONFIGS } from '../../new/transformer/schema';
interface Props {
  existingConfig?: GenerateCardNumber;
  onSubmit(config: GenerateCardNumber): void;
  isReadonly?: boolean;
}

export default function GenerateCardNumberForm(props: Props): ReactElement {
  const { existingConfig, onSubmit, isReadonly } = props;

  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver(TRANSFORMER_SCHEMA_CONFIGS.generateCardNumberConfig),
    defaultValues: {
      validLuhn: existingConfig?.validLuhn ?? false,
    },
  });
  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <Form {...form}>
        <FormField
          control={form.control}
          name="validLuhn"
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Valid Luhn</FormLabel>
                <FormDescription className="w-[90%]">
                  Generate a 16 digit card number that passes a luhn check.
                </FormDescription>
              </div>
              <FormControl>
                <Switch
                  disabled={isReadonly}
                  checked={field.value}
                  onCheckedChange={field.onChange}
                />
              </FormControl>
            </FormItem>
          )}
        />
        <div className="flex justify-end">
          <Button
            type="button"
            disabled={isReadonly}
            onClick={(e) => {
              form.handleSubmit((values) => {
                onSubmit(
                  new GenerateCardNumber({
                    ...values,
                  })
                );
              })(e);
            }}
          >
            Save
          </Button>
        </div>
      </Form>
    </div>
  );
}
