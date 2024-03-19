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
import { TransformInt64 } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { TransformerFormProps } from './util';

interface Props extends TransformerFormProps<TransformInt64> {}

export default function TransformInt64Form(props: Props): ReactElement {
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
          name={`randomizationRangeMin`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Minimum Range Value</FormLabel>
                <FormDescription className="w-[90%]">
                  Sets a minium lower range value. This will create an
                  lowerbound around the source input value. For example, if the
                  input value is 10, and you set this value to 5, then the
                  maximum range will be 5.
                </FormDescription>
              </div>
              <FormControl>
                <div className="max-w-[180px]">
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
                </div>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={`randomizationRangeMax`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Maxiumum Range Value</FormLabel>
                <FormDescription className="w-[90%]">
                  Sets a maximum upper range value. This will create an
                  upperbound around the source input value. For example, if the
                  input value is 10, and you set this value to 5, then the
                  maximum range will be 15.
                </FormDescription>
              </div>
              <FormControl>
                <div className="max-w-[180px]">
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
                </div>
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
                  new TransformInt64({
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
