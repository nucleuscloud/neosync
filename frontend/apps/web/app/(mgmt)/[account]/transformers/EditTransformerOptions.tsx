import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
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
  Transformer,
  isSystemTransformer,
  isUserDefinedTransformer,
} from '@/shared/transformers';
import { getTransformerDataTypesString } from '@/util/util';
import {
  JobMappingTransformerForm,
  convertJobMappingTransformerToForm,
  convertTransformerConfigToForm,
} from '@/yup-validations/jobs';
import { PlainMessage } from '@bufbuild/protobuf';
import {
  GenerateCardNumber,
  GenerateCategorical,
  GenerateE164PhoneNumber,
  GenerateEmail,
  GenerateFloat64,
  GenerateGender,
  GenerateInt64,
  GenerateJavascript,
  GenerateString,
  GenerateStringPhoneNumber,
  GenerateUuid,
  JobMappingTransformer,
  TransformCharacterScramble,
  TransformE164PhoneNumber,
  TransformEmail,
  TransformFirstName,
  TransformFloat64,
  TransformFullName,
  TransformInt64,
  TransformInt64PhoneNumber,
  TransformJavascript,
  TransformLastName,
  TransformPhoneNumber,
  TransformString,
  TransformerConfig,
  TransformerSource,
} from '@neosync/sdk';
import {
  Cross2Icon,
  EyeOpenIcon,
  MixerHorizontalIcon,
  Pencil1Icon,
} from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import GenerateCardNumberForm from './Sheetforms/GenerateCardNumberForm';
import GenerateCategoricalForm from './Sheetforms/GenerateCategoricalForm';
import GenerateEmailForm from './Sheetforms/GenerateEmailForm';
import GenerateFloatForm from './Sheetforms/GenerateFloat64Form';
import GenerateGenderForm from './Sheetforms/GenerateGenderForm';
import GenerateIntForm from './Sheetforms/GenerateInt64Form';
import GenerateInternationalPhoneNumberForm from './Sheetforms/GenerateInternationalPhoneNumberForm';
import GenerateJavascriptForm from './Sheetforms/GenerateJavascriptForm';
import GenerateStringForm from './Sheetforms/GenerateRandomStringForm';
import GenerateStringPhoneNumberForm from './Sheetforms/GenerateStringPhoneNumberForm';
import GenerateUuidForm from './Sheetforms/GenerateUuidForm';
import TransformCharacterScrambleForm from './Sheetforms/TransformCharacterScrambleForm';
import TransformE164NumberForm from './Sheetforms/TransformE164PhoneNumberForm';
import TransformEmailForm from './Sheetforms/TransformEmailForm';
import TransformFirstNameForm from './Sheetforms/TransformFirstNameForm';
import TransformFloatForm from './Sheetforms/TransformFloat64Form';
import TransformFullNameForm from './Sheetforms/TransformFullNameForm';
import TransformInt64Form from './Sheetforms/TransformInt64Form';
import TransformInt64PhoneForm from './Sheetforms/TransformInt64PhoneForm';
import TransformJavascriptForm from './Sheetforms/TransformJavascriptForm';
import TransformLastNameForm from './Sheetforms/TransformLastNameForm';
import TransformStringPhoneNumberForm from './Sheetforms/TransformPhoneNumberForm';
import TransformStringForm from './Sheetforms/TransformStringForm';

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
          className="ml-auto hidden h-[36px] lg:flex"
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
            <ConfigureTransformer
              transformer={transformer}
              value={value}
              onSubmit={(newval) => {
                onSubmit(newval);
                setIsSheetOpen(false);
              }}
              isReadonly={disabled || isUserDefinedTransformer(transformer)}
            />
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}

interface ConfigureTransformerProps {
  transformer: Transformer;
  value: JobMappingTransformerForm;
  onSubmit(newValue: JobMappingTransformerForm): void;
  isReadonly: boolean;
}

function ConfigureTransformer(props: ConfigureTransformerProps): ReactElement {
  const { transformer, value, onSubmit, isReadonly } = props;

  const valueConfig = isSystemTransformer(transformer)
    ? value.config
    : convertTransformerConfigToForm(transformer.config);

  switch (transformer.source) {
    case TransformerSource.GENERATE_CARD_NUMBER:
      return (
        <GenerateCardNumberForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateCardNumber({
              ...(valueConfig.value as PlainMessage<GenerateCardNumber>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'generateCardNumberConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.GENERATE_E164_PHONE_NUMBER:
      return (
        <GenerateInternationalPhoneNumberForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateE164PhoneNumber({
              ...(valueConfig.value as PlainMessage<GenerateE164PhoneNumber>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'generateE164PhoneNumberConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.GENERATE_FLOAT64:
      return (
        <GenerateFloatForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateFloat64({
              ...(valueConfig.value as PlainMessage<GenerateFloat64>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'generateFloat64Config',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.GENERATE_GENDER:
      return (
        <GenerateGenderForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateGender({
              ...(valueConfig.value as PlainMessage<GenerateGender>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'generateGenderConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.GENERATE_INT64:
      return (
        <GenerateIntForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateInt64({
              ...(valueConfig.value as PlainMessage<GenerateInt64>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'generateInt64Config',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.GENERATE_RANDOM_STRING:
      return (
        <GenerateStringForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateString({
              ...(valueConfig.value as PlainMessage<GenerateString>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'generateStringConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.GENERATE_STRING_PHONE_NUMBER:
      return (
        <GenerateStringPhoneNumberForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateStringPhoneNumber({
              ...(valueConfig.value as PlainMessage<GenerateStringPhoneNumber>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'generateStringPhoneNumberConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.GENERATE_UUID:
      return (
        <GenerateUuidForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateUuid({
              ...(valueConfig.value as PlainMessage<GenerateUuid>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'generateUuidConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.TRANSFORM_E164_PHONE_NUMBER:
      return (
        <TransformE164NumberForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformE164PhoneNumber({
              ...(valueConfig.value as PlainMessage<TransformE164PhoneNumber>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'transformE164PhoneNumberConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.TRANSFORM_EMAIL:
      return (
        <TransformEmailForm
          isReadonly={isReadonly}
          existingConfig={TransformEmail.fromJson(valueConfig.value)}
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'transformEmailConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.GENERATE_EMAIL:
      return (
        <GenerateEmailForm
          isReadonly={isReadonly}
          existingConfig={GenerateEmail.fromJson(valueConfig.value)}
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'generateEmailConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.TRANSFORM_FIRST_NAME:
      return (
        <TransformFirstNameForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformFirstName({
              ...(valueConfig.value as PlainMessage<TransformFirstName>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'transformFirstNameConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.TRANSFORM_FLOAT64:
      return (
        <TransformFloatForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformFloat64({
              ...(valueConfig.value as PlainMessage<TransformFloat64>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'transformFloat64Config',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.TRANSFORM_FULL_NAME:
      return (
        <TransformFullNameForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformFullName({
              ...(valueConfig.value as PlainMessage<TransformFullName>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'transformFullNameConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.TRANSFORM_INT64:
      return (
        <TransformInt64Form
          isReadonly={isReadonly}
          existingConfig={
            new TransformInt64({
              ...(valueConfig.value as PlainMessage<TransformInt64>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'transformInt64Config',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.TRANSFORM_INT64_PHONE_NUMBER:
      return (
        <TransformInt64PhoneForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformInt64PhoneNumber({
              ...(valueConfig.value as PlainMessage<TransformInt64PhoneNumber>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'transformInt64PhoneNumberConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.TRANSFORM_LAST_NAME:
      return (
        <TransformLastNameForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformLastName({
              ...(valueConfig.value as PlainMessage<TransformLastName>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'transformLastNameConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.TRANSFORM_PHONE_NUMBER:
      return (
        <TransformStringPhoneNumberForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformPhoneNumber({
              ...(valueConfig.value as PlainMessage<TransformPhoneNumber>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'transformPhoneNumberConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.TRANSFORM_STRING:
      return (
        <TransformStringForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformString({
              ...(valueConfig.value as PlainMessage<TransformString>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'transformStringConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.TRANSFORM_JAVASCRIPT:
      return (
        <TransformJavascriptForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformJavascript({
              ...(valueConfig.value as PlainMessage<TransformJavascript>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'transformJavascriptConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.GENERATE_CATEGORICAL:
      return (
        <GenerateCategoricalForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateCategorical({
              ...(valueConfig.value as PlainMessage<GenerateCategorical>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'generateCategoricalConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.TRANSFORM_CHARACTER_SCRAMBLE:
      return (
        <TransformCharacterScrambleForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformCharacterScramble({
              ...(valueConfig.value as PlainMessage<TransformCharacterScramble>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'transformCharacterScrambleConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );
    case TransformerSource.GENERATE_JAVASCRIPT:
      return (
        <GenerateJavascriptForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateJavascript({
              ...(valueConfig.value as PlainMessage<GenerateJavascript>),
            })
          }
          onSubmit={(newconfig) => {
            onSubmit(
              convertJobMappingTransformerToForm(
                new JobMappingTransformer({
                  source: transformer.source,
                  config: new TransformerConfig({
                    config: {
                      case: 'generateJavascriptConfig',
                      value: newconfig,
                    },
                  }),
                })
              )
            );
          }}
        />
      );

    default:
      <div>No transformer component found</div>;
  }
  return (
    <div>
      <Alert className="border-gray-200 dark:border-gray-700 shadow-sm">
        <div className="flex flex-row items-center gap-4">
          <MixerHorizontalIcon className="h-4 w-4" />
          <AlertDescription className="text-gray-500">
            There are no additional configurations for this Transformer
          </AlertDescription>
        </div>
      </Alert>
    </div>
  );
}
