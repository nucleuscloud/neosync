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
import { SystemTransformer, UserDefinedTransformer } from '@neosync/sdk';
import {
  Cross2Icon,
  MixerHorizontalIcon,
  Pencil1Icon,
} from '@radix-ui/react-icons';
import { ReactElement, useEffect, useRef, useState } from 'react';
import GenerateCardNumberForm from './Sheetforms/GenerateCardNumberForm';
import GenerateCategoricalForm from './Sheetforms/GenerateCategoricalForm';
import GenerateE164PhoneNumberForm from './Sheetforms/GenerateE164PhoneNumberForm';
import GenerateFloatForm from './Sheetforms/GenerateFloat64Form';
import GenerateGenderForm from './Sheetforms/GenerateGenderForm';
import GenerateIntForm from './Sheetforms/GenerateInt64Form';
import GenerateStringForm from './Sheetforms/GenerateStringForm';
import GenerateStringPhoneForm from './Sheetforms/GenerateStringPhoneNumberForm';
import GenerateUuidForm from './Sheetforms/GenerateUuidForm';
import TransformE164NumberForm from './Sheetforms/TransformE164PhoneNumberForm';
import TransformEmailForm from './Sheetforms/TransformEmailForm';
import TransformFirstNameForm from './Sheetforms/TransformFirstNameForm';
import TransformFloatForm from './Sheetforms/TransformFloat64Form';
import TransformFullNameForm from './Sheetforms/TransformFullNameForm';
import TransformInt64Form from './Sheetforms/TransformInt64Form';
import TransformInt64PhoneForm from './Sheetforms/TransformInt64PhoneForm';
import TransformJavascriptForm from './Sheetforms/TransformJavascriptForm';
import TransformLastNameForm from './Sheetforms/TransformLastNameForm';
import TransformPhoneNumberForm from './Sheetforms/TransformPhoneNumberForm';
import TransformStringForm from './Sheetforms/TransformStringForm';

interface Props {
  transformer: Transformer | undefined;
  // mapping index
  index: number;
}

export default function EditTransformerOptions(props: Props): ReactElement {
  const { transformer, index } = props;

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
          disabled={!transformer || isUserDefinedTransformer(transformer)}
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
            handleTransformerForm(transformer, index, setIsSheetOpen)}
        </div>
      </SheetContent>
    </Sheet>
  );
}

function handleTransformerForm(
  transformer: Transformer,
  index?: number,
  setIsSheetOpen?: (val: boolean) => void
): ReactElement {
  switch (transformer.source) {
    case 'generate_card_number':
      return (
        <GenerateCardNumberForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'generate_e164_phone_number':
      return (
        <GenerateE164PhoneNumberForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
        />
      );
    case 'generate_float64':
      return (
        <GenerateFloatForm index={index} setIsSheetOpen={setIsSheetOpen} />
      );
    case 'generate_gender':
      return (
        <GenerateGenderForm index={index} setIsSheetOpen={setIsSheetOpen} />
      );
    case 'generate_int64':
      return <GenerateIntForm index={index} setIsSheetOpen={setIsSheetOpen} />;
    case 'generate_string':
      return (
        <GenerateStringForm index={index} setIsSheetOpen={setIsSheetOpen} />
      );
    case 'generate_string_phone':
      return (
        <GenerateStringPhoneForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
        />
      );
    case 'generate_uuid':
      return <GenerateUuidForm index={index} setIsSheetOpen={setIsSheetOpen} />;
    case 'transform_e164_phone_number':
      return (
        <TransformE164NumberForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'transform_email':
      return (
        <TransformEmailForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'transform_first_name':
      return (
        <TransformFirstNameForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'transform_float64':
      return (
        <TransformFloatForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'transform_full_name':
      return (
        <TransformFullNameForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'transform_int64':
      return (
        <TransformInt64Form
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'transform_int64_phone_number':
      return (
        <TransformInt64PhoneForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'transform_last_name':
      return (
        <TransformLastNameForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'transform_phone_number':
      return (
        <TransformPhoneNumberForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'transform_string':
      return (
        <TransformStringForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'transform_javascript':
      return (
        <TransformJavascriptForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'generate_categorical':
      return (
        <GenerateCategoricalForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
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
      (t.source == 'null' ||
        t.source == 'default' ||
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
