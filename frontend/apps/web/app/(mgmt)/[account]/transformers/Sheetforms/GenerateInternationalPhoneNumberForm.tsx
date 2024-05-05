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
import { GenerateE164PhoneNumber } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { TRANSFORMER_SCHEMA_CONFIGS } from '../../new/transformer/schema';
interface Props {
  existingConfig?: GenerateE164PhoneNumber;
  onSubmit(config: GenerateE164PhoneNumber): void;
  isReadonly?: boolean;
}

export default function GenerateInternationalPhoneNumberForm(
  props: Props
): ReactElement {
  const { existingConfig, onSubmit, isReadonly } = props;

  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver(
      TRANSFORMER_SCHEMA_CONFIGS.generateE164PhoneNumberConfig
    ),
    defaultValues: {
      min: existingConfig?.min ?? BigInt(9),
      max: existingConfig?.max ?? BigInt(15),
    },
  });
  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <Form {...form}>
        <FormField
          control={form.control}
          name="min"
          render={({ field }) => (
            <FormItem className="rounded-lg border p-3 shadow-sm">
              <div className="flex flex-row items-start justify-between">
                <div className="flex flex-col space-y-2">
                  <FormLabel>Min</FormLabel>
                  <FormDescription className="w-[90%]">
                    Set the minimum length range of the output phone number. It
                    cannot be less than 9.
                  </FormDescription>
                </div>
                <FormControl>
                  <div className=" flex flex-col items-center max-w-[180px]">
                    <Input
                      type="number"
                      {...field}
                      value={String(field.value)}
                      onChange={(event) =>
                        field.onChange(event.target.valueAsNumber)
                      }
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
          name="max"
          render={({ field }) => (
            <FormItem className="rounded-lg border p-3 shadow-sm">
              <div className="flex flex-row items-start justify-between">
                <div className="flex flex-col space-y-2">
                  <FormLabel>Max</FormLabel>
                  <FormDescription className="w-[90%]">
                    Set the maximum length range of the output phone number. It
                    cannot be greater than 15.
                  </FormDescription>
                </div>
                <FormControl>
                  <div className=" flex flex-col items-center max-w-[180px]">
                    <Input
                      type="number"
                      {...field}
                      value={String(field.value)}
                      onChange={(event) =>
                        field.onChange(event.target.valueAsNumber)
                      }
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
                  new GenerateE164PhoneNumber({
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
