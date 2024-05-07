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
import { yupResolver } from '@hookform/resolvers/yup';
import { TransformFloat64 } from '@neosync/sdk';
import { ReactElement, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { TRANSFORMER_SCHEMA_CONFIGS } from '../../new/transformer/schema';
import { TransformerFormProps } from './util';

interface Props extends TransformerFormProps<TransformFloat64> {}

export default function TransformFloat64Form(props: Props): ReactElement {
  const { existingConfig, onSubmit, isReadonly } = props;

  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver(TRANSFORMER_SCHEMA_CONFIGS.transformFloat64Config),
    defaultValues: {
      randomizationRangeMin: existingConfig?.randomizationRangeMin ?? 1,
      randomizationRangeMax: existingConfig?.randomizationRangeMax ?? 40,
    },
  });

  const min = form.watch('randomizationRangeMin');
  const max = form.watch('randomizationRangeMax');

  useEffect(() => {
    form.trigger('randomizationRangeMin');
  }, [max, form.trigger]);

  useEffect(() => {
    form.trigger('randomizationRangeMax');
  }, [min, form.trigger]);

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <Form {...form}>
        <FormField
          control={form.control}
          name={`randomizationRangeMin`}
          render={({ field }) => (
            <FormItem className="rounded-lg border p-3 shadow-sm">
              <div className="flex flex-row items-start justify-between">
                <div className="flex flex-col space-y-2">
                  <FormLabel>Relative Minimum Value</FormLabel>
                  <FormDescription className="w-[90%]">
                    Sets a relative minium lower range value. This will create
                    an lowerbound around the source input value. For example, if
                    the input value is 10, and you set this value to 5, then the
                    maximum range will be 5.
                  </FormDescription>
                </div>
                <FormControl>
                  <div className="flex flex-col items-center">
                    <Input
                      {...field}
                      className="max-w-[180px]"
                      type="number"
                      value={field.value ? field.value.toString() : 0}
                      onChange={(event) => {
                        field.onChange(event.target.valueAsNumber);
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
          name={`randomizationRangeMax`}
          render={({ field }) => (
            <FormItem className="rounded-lg border p-3 shadow-sm">
              <div className="flex flex-row items-start justify-between">
                <div className="flex flex-col space-y-2">
                  <FormLabel>Relative Maximum Range Value</FormLabel>
                  <FormDescription className="w-[90%]">
                    Sets a relative maximum upper range value. This will create
                    an upperbound around the source input value. For example, if
                    the input value is 10, and you set this value to 5, then the
                    maximum range will be 15.
                  </FormDescription>
                </div>
                <FormControl>
                  <div className="flex flex-col items-center">
                    <Input
                      {...field}
                      className="max-w-[180px]"
                      type="number"
                      value={field.value ? field.value.toString() : 0}
                      onChange={(event) => {
                        field.onChange(event.target.valueAsNumber);
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
                  new TransformFloat64({
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
