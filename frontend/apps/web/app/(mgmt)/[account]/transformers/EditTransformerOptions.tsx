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
import {
  Transformer,
  isSystemTransformer,
  isUserDefinedTransformer,
} from '@/shared/transformers';
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
  EyeOpenIcon,
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
  transformer: Transformer;
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
          disabled={disabled}
          onClick={() => setIsSheetOpen(true)}
          className="ml-auto hidden h-[36px] lg:flex"
        >
          {isUserDefinedTransformer(transformer) ? (
            <EyeOpenIcon />
          ) : (
            <Pencil1Icon />
          )}
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
    case 'generate_card_number':
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
    case 'generate_e164_phone_number':
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
    case 'generate_float64':
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
    case 'generate_gender':
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
    case 'generate_int64':
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
    case 'generate_string':
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
    case 'generate_string_phone_number':
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
    case 'generate_uuid':
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
    case 'transform_e164_phone_number':
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
    case 'transform_email':
      return (
        <TransformEmailForm
          isReadonly={isReadonly}
          existingConfig={
            new TransformEmail({
              ...(valueConfig.value as PlainMessage<TransformEmail>),
            })
          }
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
    case 'transform_first_name':
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
    case 'transform_float64':
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
    case 'transform_full_name':
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
    case 'transform_int64':
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
    case 'transform_int64_phone_number':
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
    case 'transform_last_name':
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
    case 'transform_string_phone_number':
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
    case 'transform_string':
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
    case 'transform_javascript':
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
    case 'generate_categorical':
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
    case 'transform_character_scramble':
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
