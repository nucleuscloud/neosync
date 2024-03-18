import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet';
import { Transformer, isUserDefinedTransformer } from '@/shared/transformers';
import {
  JobMappingTransformerForm,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import { PlainMessage } from '@bufbuild/protobuf';
import {
  GenerateCardNumber,
  GenerateCategorical,
  GenerateE164PhoneNumber,
  GenerateFloat64,
  GenerateGender,
  GenerateInt64,
  GenerateString,
  GenerateStringPhoneNumber,
  GenerateUuid,
  JobMappingTransformer,
  SystemTransformer,
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
  UserDefinedTransformer,
} from '@neosync/sdk';
import {
  Cross2Icon,
  MixerHorizontalIcon,
  Pencil1Icon,
} from '@radix-ui/react-icons';
import { ReactElement, useEffect, useRef, useState } from 'react';
import GenerateCardNumberForm from './Sheetforms/GenerateCardNumberForm';
import GenerateCategoricalForm from './Sheetforms/GenerateCategoricalForm';
import GenerateFloatForm from './Sheetforms/GenerateFloat64Form';
import GenerateGenderForm from './Sheetforms/GenerateGenderForm';
import GenerateIntForm from './Sheetforms/GenerateInt64Form';
import GenerateInternationalPhoneNumberForm from './Sheetforms/GenerateInternationalPhoneNumberForm';
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
  transformer: Transformer | undefined;
  disabled: boolean;
  value: JobMappingTransformerForm;
  onSubmit(newValue: JobMappingTransformerForm): void;
}

export default function EditTransformerOptions(props: Props): ReactElement {
  const { transformer, disabled, value, onSubmit } = props;

  const [isSheetOpen, setIsSheetOpen] = useState(false);
  const sheetRef = useRef<HTMLDivElement | null>(null);

  // handles click outside of sheet so that it closes correctly
  useEffect(() => {
    const handleOutsideClick = (event: MouseEvent) => {
      if (
        sheetRef.current &&
        !sheetRef.current.contains(event.target as Node)
      ) {
        setIsSheetOpen(false);
      }
    };

    if (isSheetOpen) {
      document.addEventListener('mousedown', handleOutsideClick);
    }

    return () => {
      document.removeEventListener('mousedown', handleOutsideClick);
    };
  }, [isSheetOpen]);

  // since component is in a controlled state, have to manually handle closing the sheet when the user presses escape
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setIsSheetOpen!(false);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    // Clean up the event listener when the component unmounts
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, []);

  return (
    <Sheet open={isSheetOpen} onOpenChange={() => setIsSheetOpen(true)}>
      <SheetTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          // disabling this form if the transformer is user defined becuase the form is meant to load job mappings that are system transformers
          // however, that doesn't really work when the job mapping is "custom" because the config is not a system transformer config so it doens't know how to load the values
          // we need to load the custom transformer values and push them into the component, but the components expect the "form", which is the Job Mapping.
          // this would require a refactor of the lower components to not rely on the react-hook-form and instead values as props to the component itself.
          // until that is true, this needs to be disabled.
          disabled={
            !transformer || isUserDefinedTransformer(transformer) || disabled
          }
          onClick={() => setIsSheetOpen(true)}
          className="ml-auto hidden h-[36px] lg:flex"
        >
          <Pencil1Icon />
        </Button>
      </SheetTrigger>
      <SheetContent className="w-[800px]" ref={sheetRef}>
        <SheetHeader>
          <div className="flex flex-row justify-between w-full">
            <div className="flex flex-col space-y-2">
              <div className="flex flex-row gap-2">
                <SheetTitle>{transformer?.name}</SheetTitle>
                <Badge variant="outline">{transformer?.dataType}</Badge>
              </div>
              <SheetDescription>{transformer?.description}</SheetDescription>
            </div>
            <Button variant="ghost" onClick={() => setIsSheetOpen(false)}>
              <Cross2Icon className="h-4 w-4" />
            </Button>
          </div>
          <Separator />
        </SheetHeader>
        <div className="pt-8">
          {transformer &&
            handleTransformerForm(transformer, value, onSubmit, disabled)}
        </div>
      </SheetContent>
    </Sheet>
  );
}

function handleTransformerForm(
  transformer: Transformer,
  value: JobMappingTransformerForm,
  onSubmit: (newValue: JobMappingTransformerForm) => void,
  isReadonly: boolean
): ReactElement {
  const index = -1;
  const setIsSheetOpen = () => undefined;
  switch (transformer.source) {
    case 'generate_card_number':
      return (
        <GenerateCardNumberForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateCardNumber({
              ...(value.config.value as PlainMessage<GenerateCardNumber>),
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
    case 'generate_international_phone_number':
      return (
        <GenerateInternationalPhoneNumberForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateE164PhoneNumber({
              ...(value.config.value as PlainMessage<GenerateE164PhoneNumber>),
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
    case 'generate_float64':
      return (
        <GenerateFloatForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateFloat64({
              ...(value.config.value as PlainMessage<GenerateFloat64>),
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
    case 'generate_gender':
      return (
        <GenerateGenderForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateGender({
              ...(value.config.value as PlainMessage<GenerateGender>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'generate_int64':
      return (
        <GenerateIntForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateInt64({
              ...(value.config.value as PlainMessage<GenerateInt64>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'generate_string':
      return (
        <GenerateStringForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateString({
              ...(value.config.value as PlainMessage<GenerateString>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'generate_string_phone_number':
      return (
        <GenerateStringPhoneNumberForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateStringPhoneNumber({
              ...(value.config
                .value as PlainMessage<GenerateStringPhoneNumber>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'generate_uuid':
      return (
        <GenerateUuidForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateUuid({
              ...(value.config.value as PlainMessage<GenerateUuid>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'transform_e164_phone_number':
      return (
        <TransformE164NumberForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformE164PhoneNumber({
              ...(value.config.value as PlainMessage<TransformE164PhoneNumber>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'transform_email':
      return (
        <TransformEmailForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformEmail({
              ...(value.config.value as PlainMessage<TransformEmail>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'transform_first_name':
      return (
        <TransformFirstNameForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformFirstName({
              ...(value.config.value as PlainMessage<TransformFirstName>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'transform_float64':
      return (
        <TransformFloatForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformFloat64({
              ...(value.config.value as PlainMessage<TransformFloat64>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'transform_full_name':
      return (
        <TransformFullNameForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformFullName({
              ...(value.config.value as PlainMessage<TransformFullName>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'transform_int64':
      return (
        <TransformInt64Form
          isReadonly={isReadonly}
          existingConfig={
            new TransformInt64({
              ...(value.config.value as PlainMessage<TransformInt64>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'transform_int64_phone_number':
      return (
        <TransformInt64PhoneForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformInt64PhoneNumber({
              ...(value.config
                .value as PlainMessage<TransformInt64PhoneNumber>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'transform_last_name':
      return (
        <TransformLastNameForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformLastName({
              ...(value.config.value as PlainMessage<TransformLastName>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'transform_string_phone_number':
      return (
        <TransformStringPhoneNumberForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformPhoneNumber({
              ...(value.config.value as PlainMessage<TransformPhoneNumber>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'transform_string':
      return (
        <TransformStringForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformString({
              ...(value.config.value as PlainMessage<TransformString>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'transform_javascript':
      return (
        <TransformJavascriptForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformJavascript({
              ...(value.config.value as PlainMessage<TransformJavascript>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'generate_categorical':
      return (
        <GenerateCategoricalForm
          isReadonly={isReadonly}
          existingConfig={
            new GenerateCategorical({
              ...(value.config.value as PlainMessage<GenerateCategorical>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );
    case 'transform_character_scramble':
      return (
        <TransformCharacterScrambleForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformCharacterScramble({
              ...(value.config
                .value as PlainMessage<TransformCharacterScramble>),
            })
          }
          onSubmit={(newconfig) => {
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
            );
          }}
        />
      );

    default:
      <div>No transformer component found</div>;
  }
  return (
    <div>
      <Alert className="border-gray-200 shadow-sm">
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

export function filterInputFreeSystemTransformers(
  transformers: SystemTransformer[]
): SystemTransformer[] {
  return transformers.filter(
    (t) =>
      t.source !== 'passthrough' &&
      (t.source === 'null' ||
        t.source === 'default' ||
        t.source.startsWith('generate_'))
  );
}

export function filterInputFreeUdfTransformers(
  udfTransformers: UserDefinedTransformer[],
  systemTransformers: SystemTransformer[]
): UserDefinedTransformer[] {
  const sysMap = new Map(
    filterInputFreeSystemTransformers(systemTransformers).map((t) => [
      t.source,
      t,
    ])
  );
  return udfTransformers.filter((t) => sysMap.has(t.source));
}
