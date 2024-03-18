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
import { GenerateE164PhoneNumber } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
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
    // resolver: yupResolver(
    //   TRANSFORMER_SCHEMA_CONFIGS.generateE164PhoneNumberConfig
    // ),
    defaultValues: {
      ...existingConfig,
    },
  });
  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <Form {...form}>
        <FormField
          control={form.control}
          name="min"
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Valid Luhn</FormLabel>
                <FormDescription className="w-[90%]">
                  Set the minimum length range of the output phone number. It
                  cannot be less than 9.
                </FormDescription>
              </div>
              <FormControl>
                <div className="max-w-[180px]">
                  <Input
                    type="number"
                    {...field}
                    value={String(field.value)}
                    onChange={(event) =>
                      field.onChange(event.target.valueAsNumber)
                    }
                  />
                </div>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="max"
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Valid Luhn</FormLabel>
                <FormDescription className="w-[90%]">
                  Set the maximum length range of the output phone number. It
                  cannot be greater than 15.
                </FormDescription>
              </div>
              <FormControl>
                <div className="max-w-[180px]">
                  <Input
                    type="number"
                    {...field}
                    value={String(field.value)}
                    onChange={(event) =>
                      field.onChange(event.target.valueAsNumber)
                    }
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

  // const fc = useFormContext();

  // const minVal = fc.getValues(`mappings.${index}.transformer.config.value.min`);
  // const maxVal = fc.getValues(`mappings.${index}.transformer.config.value.max`);
  // const [min, setMin] = useState<number>(minVal);
  // const [max, setMax] = useState<number>(maxVal);
  // const [disableSave, setDisableSave] = useState<boolean>(false);
  // const [minError, setMinError] = useState<string>('');
  // const [maxError, setMaxError] = useState<string>('');

  // const handleSubmit = () => {
  //   fc.setValue(
  //     `mappings.${index}.transformer.config.value`,
  //     new GenerateE164PhoneNumber({
  //       min: BigInt(min),
  //       max: BigInt(max),
  //     }),
  //     {
  //       shouldValidate: false,
  //     }
  //   );
  //   setIsSheetOpen!(false);
  // };

  // const handleSettingMinRange = (value: number) => {
  //   if (value < 9 || value > max) {
  //     setMinError(
  //       'Minimum length cannot be less than 9 or greater than the max length'
  //     );
  //     setMin(value);
  //     setDisableSave(true);
  //   } else {
  //     setMinError('');
  //     setDisableSave(false);
  //     setMin(value);
  //   }
  // };
  // const handleSettingMaxRange = (value: number) => {
  //   if (value > 15 || value < min) {
  //     setMaxError(
  //       'Maximum length cannot be greater than 15 or less than the min length'
  //     );
  //     setMax(value);
  //     setDisableSave(true);
  //   } else {
  //     setMaxError('');
  //     setDisableSave(false);
  //     setMax(value);
  //   }
  // };

  // return (
  //   <div className="flex flex-col w-full space-y-4 pt-4">
  //     <FormField
  //       name={`mappings.${index}.transformer.config.value.min`}
  //       render={() => (
  //         <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
  //           <div className="space-y-0.5">
  //             <FormLabel>Minimum Length</FormLabel>
  //             <FormDescription>
  //               Set the minimum length range of the output phone number. It
  //               cannot be less than 9.
  //             </FormDescription>
  //           </div>
  //           <FormControl>
  //             <div className="max-w-[180px]">
  //               <Input
  //                 value={String(min)}
  //                 onChange={(event) =>
  //                   handleSettingMinRange(Number(event.target.value))
  //                 }
  //               />
  //               <FormError errorMessage={minError} />
  //             </div>
  //           </FormControl>
  //         </FormItem>
  //       )}
  //     />
  //     <FormField
  //       name={`mappings.${index}.transformer.config.value.max`}
  //       render={() => (
  //         <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
  //           <div className="space-y-0.5">
  //             <FormLabel>Maximum Length</FormLabel>
  //             <FormDescription>
  //               Set the maximum length range of the output phone number. It
  //               cannot be greater than 15.
  //             </FormDescription>
  //           </div>
  //           <FormControl>
  //             <div className="max-w-[180px]">
  //               <Input
  //                 value={String(max)}
  //                 onChange={(event) =>
  //                   handleSettingMaxRange(Number(event.target.value))
  //                 }
  //               />
  //               <FormError errorMessage={maxError} />
  //             </div>
  //           </FormControl>
  //         </FormItem>
  //       )}
  //     />
  //     <div className="flex justify-end">
  //       <Button type="button" onClick={handleSubmit} disabled={disableSave}>
  //         Save
  //       </Button>
  //     </div>
  //   </div>
  // );
}
