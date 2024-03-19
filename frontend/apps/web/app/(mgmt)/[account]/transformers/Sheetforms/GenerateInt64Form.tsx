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
import { GenerateInt64 } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { TransformerFormProps } from './util';

interface Props extends TransformerFormProps<GenerateInt64> {}

export default function GenerateInt64Form(props: Props): ReactElement {
  const { existingConfig, onSubmit, isReadonly } = props;

  const form = useForm({
    mode: 'onChange',
    defaultValues: {
      ...existingConfig,
    },
  });

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <Form {...form}>
        <FormField
          control={form.control}
          name={`randomizeSign`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Randomize Sign</FormLabel>
                <FormDescription>
                  Randomly the generated value between positive and negative. By
                  default, it generates a positive number.
                </FormDescription>
              </div>
              <FormControl>
                <Switch
                  checked={field.value}
                  onCheckedChange={field.onChange}
                  disabled={isReadonly}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={`min`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Minimum Value</FormLabel>
                <FormDescription>
                  Sets a minimum range for generated int64 value. This can be
                  negative as well.
                </FormDescription>
              </div>
              <FormControl>
                <Input
                  {...field}
                  type="number"
                  className="max-w-[180px]"
                  value={field.value ? field.value.toString() : 0}
                  onChange={(event) => {
                    field.onChange(BigInt(event.target.valueAsNumber));
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
                  Sets a maximum range for generated int64 value. This can be
                  negative as well.
                </FormDescription>
              </div>
              <FormControl>
                <Input
                  {...field}
                  type="number"
                  className="max-w-[180px]"
                  value={field.value ? field.value.toString() : 0}
                  onChange={(event) => {
                    field.onChange(BigInt(event.target.valueAsNumber));
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
                  new GenerateInt64({
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
