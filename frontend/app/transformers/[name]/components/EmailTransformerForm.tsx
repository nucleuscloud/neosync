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
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm, useFormContext } from 'react-hook-form';
import * as Yup from 'yup';

const EMAIL_FORM_SCHEMA = Yup.object({
  value: Yup.string().required(),
  preserve_length: Yup.bool().required(),
  preserve_domain: Yup.bool().required(),
});

type FormValues = Yup.InferType<typeof EMAIL_FORM_SCHEMA>;

interface Props {
  transformer: Transformer;
  index?: number;
}

export default function EmailTransformerForm(props: Props): ReactElement {
  const { transformer, index } = props;

  const form = useForm<FormValues>({
    resolver: yupResolver(EMAIL_FORM_SCHEMA),
    defaultValues: {
      value: transformer.value ?? '',
      preserve_length: false,
      preserve_domain: false,
    },
  });

  const fc = useFormContext();

  const onSubmit = (values: FormValues) => {
    fc.setValue(`mappings.${index}.transformer.config`, values, {
      shouldValidate: false,
    });
  };

  return (
    <div className="w-full">
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <FormField
            control={form.control}
            name={'preserve_length'}
            render={({ field }) => (
              <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                <div className="space-y-0.5">
                  <FormLabel>Preserve Length</FormLabel>
                  <FormDescription>
                    Set the length of the output email to be the same as the
                    input
                  </FormDescription>
                </div>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name={'preserve_domain'}
            render={({ field }) => (
              <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
                <div className="space-y-0.5">
                  <FormLabel>Preserve Domain</FormLabel>
                  <FormDescription>
                    Set the length of the output email to be the same as the
                    input
                  </FormDescription>
                </div>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </FormItem>
            )}
          />
          <div className="flex justify-end">
            <Button type="submit">Submit</Button>
          </div>
        </form>
      </Form>
    </div>
  );
}
