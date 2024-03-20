'use client';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { yupResolver } from '@hookform/resolvers/yup';
import { GenerateFloat64 } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { TRANSFORMER_SCHEMA_CONFIGS } from '../../new/transformer/schema';
import { TransformerFormProps, setBigIntOrOld } from './util';
interface Props extends TransformerFormProps<GenerateFloat64> {}

export default function GenerateFloat64Form(props: Props): ReactElement {
  const { existingConfig, onSubmit, isReadonly } = props;

  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver(TRANSFORMER_SCHEMA_CONFIGS.generateFloat64Config),
    defaultValues: {
      precision: existingConfig?.precision ?? BigInt(0),
      min: existingConfig?.min ?? BigInt(0),
      max: existingConfig?.max ?? BigInt(40),
    },
  });

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <Form {...form}>
        <FormField
          control={form.control}
          name={'randomizeSign'}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0 z-10">
                <FormLabel>Randomize Sign</FormLabel>
                <FormDescription>
                  Randomly sets a sign to the generated float64 value. By
                  default, it generates a positive number.
                </FormDescription>
              </div>
              <FormControl>
                <Switch
                  name={field.name}
                  disabled={isReadonly}
                  checked={field.value}
                  onCheckedChange={field.onChange}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={'min'}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Minimum Value</FormLabel>
                <FormDescription>
                  Sets a minimum range for generated float64 value. This can be
                  negative as well.
                </FormDescription>
              </div>
              <FormControl>
                <Input
                  {...field}
                  className="max-w-[180px]"
                  type="number"
                  value={field.value ? field.value.toString() : 0}
                  onChange={(event) => {
                    field.onChange(
                      setBigIntOrOld(event.target.valueAsNumber, field.value)
                    );
                  }}
                  disabled={isReadonly}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={`max`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Maxiumum Value</FormLabel>
                <FormDescription>
                  Sets a maximum range for generated float64 value. This can be
                  negative as well.
                </FormDescription>
              </div>
              <FormControl>
                <Input
                  {...field}
                  className="max-w-[180px]"
                  type="number"
                  value={field.value ? field.value.toString() : 0}
                  onChange={(event) => {
                    field.onChange(
                      setBigIntOrOld(event.target.valueAsNumber, field.value)
                    );
                  }}
                  disabled={isReadonly}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={`precision`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Precision</FormLabel>
                <FormDescription>
                  Sets the precision for the entire float64 value, not just the
                  decimals. For example. a precision of 4 would update a float64
                  value of 23.567 to 23.56.
                </FormDescription>
              </div>

              <FormControl>
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
              </FormControl>
              <FormMessage />
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
                  new GenerateFloat64({
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
