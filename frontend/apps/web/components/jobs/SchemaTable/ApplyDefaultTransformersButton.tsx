'use client';

import ButtonText from '@/components/ButtonText';
import ConfirmationDialog from '@/components/ConfirmationDialog';
import SwitchCard from '@/components/switches/SwitchCard';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import { DefaultTransformerFormValues } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import { ReactElement } from 'react';
import { useForm, UseFormReturn } from 'react-hook-form';
interface Props {
  isDisabled: boolean;
  onClick(override: boolean): void;
}

export default function ApplyDefaultTransformersButton(
  props: Props
): ReactElement {
  const { isDisabled, onClick } = props;
  const form = useForm<DefaultTransformerFormValues>({
    resolver: yupResolver(DefaultTransformerFormValues),
    defaultValues: {
      overrideTransformers: false,
    },
  });
  return (
    <ConfirmationDialog
      trigger={
        <Button variant="outline" type="button" disabled={isDisabled}>
          <ButtonText text="Apply Default Transformers" />
        </Button>
      }
      headerText="Apply Default Transformers?"
      description="This setting will apply the 'Passthrough' Transformer to every column that is not Generated, while applying the 'Use Column Default' Transformer to all Generated (non-Identity)columns."
      body={<FormBody form={form} />}
      containerClassName="max-w-xl"
      onConfirm={() => {
        const override = form.getValues('overrideTransformers');
        onClick(override);
      }}
    />
  );
}

interface FormBodyProps {
  form: UseFormReturn<DefaultTransformerFormValues>;
}
function FormBody(props: FormBodyProps): ReactElement {
  const { form } = props;
  return (
    <div>
      <Form {...form}>
        <form className="space-y-8">
          <FormField
            control={form.control}
            name="overrideTransformers"
            render={({ field }) => (
              <FormItem>
                <FormControl>
                  <SwitchCard
                    isChecked={field.value}
                    onCheckedChange={field.onChange}
                    title="Override Mapped Transformers"
                    description="Do you want to overwrite the Transformers you have already mapped?"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </form>
      </Form>
    </div>
  );
}
