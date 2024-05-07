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
import { GenerateInt64 } from '@neosync/sdk';
import { ReactElement, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { TRANSFORMER_SCHEMA_CONFIGS } from '../../new/transformer/schema';
import { TransformerFormProps, setBigIntOrOld } from './util';

interface Props extends TransformerFormProps<GenerateInt64> {}

export default function GenerateInt64Form(props: Props): ReactElement {
  const { existingConfig, onSubmit, isReadonly } = props;

  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver(TRANSFORMER_SCHEMA_CONFIGS.generateInt64Config),
    defaultValues: {
      randomizeSign: existingConfig?.randomizeSign ?? false,
      min: existingConfig?.min ?? BigInt(0),
      max: existingConfig?.max ?? BigInt(40),
    },
  });

  /* this handles triggering the validations on the min and max if each other changes. For ex. the min must be less than the max, if the min > max, and then i update to max > min, the min would still show an error
  so we use a watch and trigger to revalidate it.
 */

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
      <Form {...form}>
        <FormField
          control={form.control}
          name={`randomizeSign`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Randomize Sign</FormLabel>
                <FormDescription>
                  {`After the value has been generated, will randomly flip the sign. This may cause the generated value to be out of the defined min/max range.
                  If the min/max is 20-40, the value may be in the following ranges: 20 <= x <= 40 and -40 <= x <= -20`}
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
            <FormItem className="rounded-lg border p-3 shadow-sm">
              <div className="flex flex-row items-start justify-between">
                <div className="flex flex-col space-y-2">
                  <FormLabel>Minimum Value</FormLabel>
                  <FormDescription>
                    Sets a minimum range for generated int64 value.
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
                          setBigIntOrOld(
                            event.target.valueAsNumber,
                            field.value
                          )
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
                  <FormLabel>Maximum Value</FormLabel>
                  <FormDescription>
                    Sets a maximum range for generated int64 value.
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
                          setBigIntOrOld(
                            event.target.valueAsNumber,
                            field.value
                          )
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
