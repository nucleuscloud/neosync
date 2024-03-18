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
import { GenerateString } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { TransformerFormProps } from './util';
interface Props extends TransformerFormProps<GenerateString> {}

export default function GenerateStringForm(props: Props): ReactElement {
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
          name={`min`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Minimum Length</FormLabel>
                <FormDescription>
                  Set the minimum length range of the output string.
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
                  />
                </div>
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
                <FormLabel>Maximum Length</FormLabel>
                <FormDescription>
                  Set the maximum length range of the output string.
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
                  new GenerateString({
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
