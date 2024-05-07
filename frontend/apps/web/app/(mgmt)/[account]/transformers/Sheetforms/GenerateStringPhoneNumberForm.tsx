'use client';
import { Button } from '@/components/ui/button';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { yupResolver } from '@hookform/resolvers/yup';
import { GenerateStringPhoneNumber } from '@neosync/sdk';
import { ReactElement, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { TRANSFORMER_SCHEMA_CONFIGS } from '../../new/transformer/schema';
import { TransformerFormProps, setBigIntOrOld } from './util';
interface Props extends TransformerFormProps<GenerateStringPhoneNumber> {}

export default function GenerateStringPhoneNumberForm(
  props: Props
): ReactElement {
  const { existingConfig, onSubmit, isReadonly } = props;

  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver(
      TRANSFORMER_SCHEMA_CONFIGS.generateStringPhoneNumberConfig
    ),
    defaultValues: {
      min: existingConfig?.min ?? BigInt(8),
      max: existingConfig?.max ?? BigInt(15),
    },
  });

  const min = form.watch('min');
  const max = form.watch('max');

  useEffect(() => {
    form.trigger('min');
  }, [max, form.trigger]);

  useEffect(() => {
    form.trigger('max');
  }, [min, form.trigger]);

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        control={form.control}
        name={`min`}
        render={({ field }) => (
          <FormItem className="rounded-lg border p-3 shadow-sm">
            <div className="flex flex-row items-start justify-between">
              <div className="flex flex-col space-y-2">
                <FormLabel>Minimum Length</FormLabel>
                <FormDescription>
                  Set the minimum length range of the output phone number.
                </FormDescription>
              </div>
              <FormControl>
                <div className="flex flex-col items-center">
                  <Input
                    {...field}
                    type="number"
                    className="max-w-[180px]"
                    value={field.value ? field.value.toString() : 0}
                    onChange={(event) => {
                      field.onChange(
                        setBigIntOrOld(event.target.valueAsNumber, field.value)
                      );
                    }}
                    disabled={isReadonly}
                  />
                  <FormMessage />
                </div>
              </FormControl>
            </div>
          </FormItem>
        )}
      />
      <FormField
        control={form.control}
        name={`max`}
        render={({ field }) => (
          <FormItem className="rounded-lg border p-3 shadow-sm">
            <div className="flex flex-row items-start justify-between">
              <div className="flex flex-col space-y-2">
                <FormLabel>Maximum Length</FormLabel>
                <FormDescription>
                  Set the maximum length range of the output phone number.
                </FormDescription>
              </div>
              <FormControl>
                <div className="flex flex-col items-center">
                  <Input
                    {...field}
                    type="number"
                    className="max-w-[180px]"
                    value={field.value ? field.value.toString() : 0}
                    onChange={(event) => {
                      field.onChange(
                        setBigIntOrOld(event.target.valueAsNumber, field.value)
                      );
                    }}
                    disabled={isReadonly}
                  />
                  <FormMessage />
                </div>
              </FormControl>
            </div>
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
                new GenerateStringPhoneNumber({
                  ...values,
                })
              );
            })(e);
          }}
        >
          Save
        </Button>
      </div>
    </div>
  );
}
