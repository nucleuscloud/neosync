import { useAccount } from '@/components/providers/account-provider';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem } from '@/components/ui/form';
import { Separator } from '@/components/ui/separator';
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet';
import {
  isSystemTransformer,
  isUserDefinedTransformer,
  Transformer,
} from '@/shared/transformers';
import { getTransformerDataTypesString } from '@/util/util';
import {
  convertTransformerConfigSchemaToTransformerConfig,
  convertTransformerConfigToForm,
  JobMappingTransformerForm,
} from '@/yup-validations/jobs';
import { useMutation } from '@connectrpc/connect-query';
import { Passthrough, TransformerConfig } from '@neosync/sdk';
import { validateUserJavascriptCode } from '@neosync/sdk/connectquery';
import { Cross2Icon, EyeOpenIcon, Pencil1Icon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';
import {
  EditJobMappingTransformerConfigFormContext,
  EditJobMappingTransformerConfigFormValues,
} from '../new/transformer/schema';
import { UserDefinedTransformerForm } from '../new/transformer/UserDefinedTransformerForms/UserDefinedTransformerForm';

interface Props {
  transformer: Transformer;
  disabled: boolean;
  value: JobMappingTransformerForm;
  onSubmit(newValue: JobMappingTransformerForm): void;
}

// Note: this has issues with re-rendering due to being embedded within the tanstack table.
// This will cause the sheet to close when the user clicks back onto the page.
// This is partially solved by memoizing the tanstack columns, but any time the columns need to re-render, this sheet will close if it's open.`
export default function EditTransformerOptions(props: Props): ReactElement {
  const { transformer, disabled, value, onSubmit } = props;
  const [isSheetOpen, setIsSheetOpen] = useState(false);

  return (
    <Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
      <SheetTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          disabled={disabled}
          className="hidden h-[36px] lg:flex"
          type="button"
        >
          {isUserDefinedTransformer(transformer) ? (
            <EyeOpenIcon />
          ) : (
            <Pencil1Icon />
          )}
        </Button>
      </SheetTrigger>
      <SheetContent className="w-[800px]">
        <SheetHeader>
          <div className="flex flex-row w-full">
            <div className="flex flex-col space-y-2 w-full">
              <div className="flex flex-row justify-between items-center">
                <div className="flex flex-row gap-2">
                  <SheetTitle>{transformer.name}</SheetTitle>
                  <Badge variant="outline">
                    {getTransformerDataTypesString(transformer.dataTypes)}
                  </Badge>
                </div>
                <SheetClose>
                  <Cross2Icon className="h-4 w-4" />
                  <span className="sr-only">Close</span>
                </SheetClose>
              </div>
              <SheetDescription>{transformer?.description}</SheetDescription>
            </div>
          </div>
          <Separator />
        </SheetHeader>
        <div className="pt-8">
          {transformer && (
            <EditTransformerConfig
              value={
                isSystemTransformer(transformer)
                  ? convertTransformerConfigSchemaToTransformerConfig(
                      value.config
                    )
                  : (transformer.config ??
                    new TransformerConfig({
                      config: {
                        case: 'passthroughConfig',
                        value: new Passthrough(),
                      },
                    }))
              }
              onSubmit={(newval) => {
                onSubmit({
                  ...value,
                  config: convertTransformerConfigToForm(newval),
                });
                setIsSheetOpen(false);
              }}
              isDisabled={disabled || isUserDefinedTransformer(transformer)}
            />
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}

interface EditTransformerConfigProps {
  value: TransformerConfig;
  onSubmit(newValue: TransformerConfig): void;
  isDisabled: boolean;
}

function EditTransformerConfig(
  props: EditTransformerConfigProps
): ReactElement {
  const { value, onSubmit, isDisabled } = props;

  const { account } = useAccount();
  const { mutateAsync: isJavascriptCodeValid } = useMutation(
    validateUserJavascriptCode
  );

  const form = useForm<
    EditJobMappingTransformerConfigFormValues,
    EditJobMappingTransformerConfigFormContext
  >({
    mode: 'onChange',
    defaultValues: { config: convertTransformerConfigToForm(value) },
    context: {
      accountId: account?.id ?? '',
      isUserJavascriptCodeValid: isJavascriptCodeValid,
    },
  });

  return (
    <Form {...form}>
      <div className="flex flex-col gap-8">
        <FormField
          control={form.control}
          name="config"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <UserDefinedTransformerForm
                  value={convertTransformerConfigSchemaToTransformerConfig(
                    field.value
                  )}
                  setValue={(newValue) =>
                    field.onChange(convertTransformerConfigToForm(newValue))
                  }
                  disabled={isDisabled}
                  errors={form.formState.errors}
                />
              </FormControl>
            </FormItem>
          )}
        />
        <div className="flex justify-end">
          <Button
            type="button"
            disabled={isDisabled || !form.formState.isValid}
            onClick={(e) => {
              form.handleSubmit((formValues) => {
                onSubmit(
                  convertTransformerConfigSchemaToTransformerConfig(
                    formValues.config
                  )
                );
              })(e);
            }}
          >
            Save
          </Button>
        </div>
      </div>
    </Form>
  );

  // switch (transformer.source) {
  //   case TransformerSource.GENERATE_CARD_NUMBER:
  //     return (
  //       <GenerateCardNumberForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new GenerateCardNumber({
  //             ...(valueConfig.value as PlainMessage<GenerateCardNumber>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'generateCardNumberConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.GENERATE_E164_PHONE_NUMBER:
  //     return (
  //       <GenerateInternationalPhoneNumberForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new GenerateE164PhoneNumber({
  //             ...(valueConfig.value as PlainMessage<GenerateE164PhoneNumber>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'generateE164PhoneNumberConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.GENERATE_FLOAT64:
  //     return (
  //       <GenerateFloatForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new GenerateFloat64({
  //             ...(valueConfig.value as PlainMessage<GenerateFloat64>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'generateFloat64Config',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.GENERATE_GENDER:
  //     return (
  //       <GenerateGenderForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new GenerateGender({
  //             ...(valueConfig.value as PlainMessage<GenerateGender>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'generateGenderConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.GENERATE_INT64:
  //     return (
  //       <GenerateIntForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new GenerateInt64({
  //             ...(valueConfig.value as PlainMessage<GenerateInt64>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'generateInt64Config',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.GENERATE_RANDOM_STRING:
  //     return (
  //       <GenerateStringForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new GenerateString({
  //             ...(valueConfig.value as PlainMessage<GenerateString>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'generateStringConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.GENERATE_STRING_PHONE_NUMBER:
  //     return (
  //       <GenerateStringPhoneNumberForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new GenerateStringPhoneNumber({
  //             ...(valueConfig.value as PlainMessage<GenerateStringPhoneNumber>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'generateStringPhoneNumberConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.GENERATE_STATE:
  //     return (
  //       <GenerateStateForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new GenerateState({
  //             ...(valueConfig.value as PlainMessage<GenerateState>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'generateStateConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.GENERATE_UUID:
  //     return (
  //       <GenerateUuidForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new GenerateUuid({
  //             ...(valueConfig.value as PlainMessage<GenerateUuid>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'generateUuidConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.TRANSFORM_E164_PHONE_NUMBER:
  //     return (
  //       <TransformE164NumberForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new TransformE164PhoneNumber({
  //             ...(valueConfig.value as PlainMessage<TransformE164PhoneNumber>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'transformE164PhoneNumberConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.TRANSFORM_EMAIL:
  //     return (
  //       <TransformEmailForm
  //         isReadonly={isReadonly}
  //         existingConfig={TransformEmail.fromJson(valueConfig.value)}
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'transformEmailConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.GENERATE_EMAIL:
  //     return (
  //       <GenerateEmailForm
  //         isReadonly={isReadonly}
  //         existingConfig={GenerateEmail.fromJson(valueConfig.value)}
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'generateEmailConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.TRANSFORM_FIRST_NAME:
  //     return (
  //       <TransformFirstNameForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new TransformFirstName({
  //             ...(valueConfig.value as PlainMessage<TransformFirstName>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'transformFirstNameConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.TRANSFORM_FLOAT64:
  //     return (
  //       <TransformFloatForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new TransformFloat64({
  //             ...(valueConfig.value as PlainMessage<TransformFloat64>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'transformFloat64Config',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.TRANSFORM_FULL_NAME:
  //     return (
  //       <TransformFullNameForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new TransformFullName({
  //             ...(valueConfig.value as PlainMessage<TransformFullName>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'transformFullNameConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.TRANSFORM_INT64:
  //     return (
  //       <TransformInt64Form
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new TransformInt64({
  //             ...(valueConfig.value as PlainMessage<TransformInt64>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'transformInt64Config',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.TRANSFORM_INT64_PHONE_NUMBER:
  //     return (
  //       <TransformInt64PhoneForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new TransformInt64PhoneNumber({
  //             ...(valueConfig.value as PlainMessage<TransformInt64PhoneNumber>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'transformInt64PhoneNumberConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.TRANSFORM_LAST_NAME:
  //     return (
  //       <TransformLastNameForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new TransformLastName({
  //             ...(valueConfig.value as PlainMessage<TransformLastName>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'transformLastNameConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.TRANSFORM_PHONE_NUMBER:
  //     return (
  //       <TransformStringPhoneNumberForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new TransformPhoneNumber({
  //             ...(valueConfig.value as PlainMessage<TransformPhoneNumber>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'transformPhoneNumberConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.TRANSFORM_STRING:
  //     return (
  //       <TransformStringForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new TransformString({
  //             ...(valueConfig.value as PlainMessage<TransformString>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'transformStringConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.TRANSFORM_JAVASCRIPT:
  //     return (
  //       <TransformJavascriptForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new TransformJavascript({
  //             ...(valueConfig.value as PlainMessage<TransformJavascript>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'transformJavascriptConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.GENERATE_CATEGORICAL:
  //     return (
  //       <GenerateCategoricalForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new GenerateCategorical({
  //             ...(valueConfig.value as PlainMessage<GenerateCategorical>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'generateCategoricalConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.TRANSFORM_CHARACTER_SCRAMBLE:
  //     return (
  //       <TransformCharacterScrambleForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new TransformCharacterScramble({
  //             ...(valueConfig.value as PlainMessage<TransformCharacterScramble>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'transformCharacterScrambleConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );
  //   case TransformerSource.GENERATE_JAVASCRIPT:
  //     return (
  //       <GenerateJavascriptForm
  //         isReadonly={isReadonly}
  //         existingConfig={
  //           new GenerateJavascript({
  //             ...(valueConfig.value as PlainMessage<GenerateJavascript>),
  //           })
  //         }
  //         onSubmit={(newconfig) => {
  //           onSubmit(
  //             convertJobMappingTransformerToForm(
  //               new JobMappingTransformer({
  //                 source: transformer.source,
  //                 config: new TransformerConfig({
  //                   config: {
  //                     case: 'generateJavascriptConfig',
  //                     value: newconfig,
  //                   },
  //                 }),
  //               })
  //             )
  //           );
  //         }}
  //       />
  //     );

  //   default:
  //     <div>No transformer component found</div>;
  // }
  // return (
  //   <div>
  //     <Alert className="border-gray-200 dark:border-gray-700 shadow-sm">
  //       <div className="flex flex-row items-center gap-4">
  //         <MixerHorizontalIcon className="h-4 w-4" />
  //         <AlertDescription className="text-gray-500">
  //           There are no additional configurations for this Transformer
  //         </AlertDescription>
  //       </div>
  //     </Alert>
  //   </div>
  // );
}
