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
import { TransformEmail } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { TransformerFormProps } from './util';
interface Props extends TransformerFormProps<TransformEmail> {}

export default function TransformEmailForm(props: Props): ReactElement {
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
          name={`preserveLength`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Preserve Length</FormLabel>
                <FormDescription className="w-[90%]">
                  Set the length of the output email to be the same as the input
                </FormDescription>
              </div>
              <FormControl>
                <Switch
                  name={field.name}
                  checked={field.value}
                  disabled={isReadonly}
                  onCheckedChange={field.onChange}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={`preserveDomain`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Preserve Domain</FormLabel>
                <FormDescription className="w-[90%]">
                  Preserve the input domain including top level domain to the
                  output value. For ex. if the input is john@gmail.com, the
                  output will be ij23o@gmail.com
                </FormDescription>
              </div>
              <FormControl>
                <Switch
                  name={field.name}
                  checked={field.value}
                  disabled={isReadonly}
                  onCheckedChange={field.onChange}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={`excludedDomains`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm gap-4 ">
              <div className="space-y-0.5">
                <FormLabel>Excluded Domains</FormLabel>
                <FormDescription>
                  Provide a list of comma-separated domains that you want to be
                  excluded from the transformer. Do not provide an @ with the
                  domains.{' '}
                </FormDescription>
              </div>
              <FormControl>
                <div className="min-w-[300px]">
                  <Input
                    {...field}
                    disabled={isReadonly}
                    type="string"
                    className="min-w-[300px]"
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
                  new TransformEmail({
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
