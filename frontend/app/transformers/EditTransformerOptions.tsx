import { TransformerWithType } from '@/components/jobs/SchemaTable/schema-table';
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
import { UserDefinedTransformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import {
  Cross2Icon,
  MixerHorizontalIcon,
  Pencil1Icon,
} from '@radix-ui/react-icons';
import { ReactElement, useEffect, useRef, useState } from 'react';
import GenerateCardNumberForm from './Sheetforms/GenerateCardNumberForm';
import GenerateE164NumberForm from './Sheetforms/GenerateE164NumberForm';
import GenerateFloatForm from './Sheetforms/GenerateFloatForm';
import GenerateGenderForm from './Sheetforms/GenerateGenderForm';
import GenerateIntForm from './Sheetforms/GenerateIntForm';
import GenerateStringForm from './Sheetforms/GenerateStringForm';
import GenerateStringPhoneForm from './Sheetforms/GenerateStringPhoneForm';
import GenerateUuidForm from './Sheetforms/GenerateUuidForm';
import TransformE164NumberForm from './Sheetforms/TransformE164NumberForm';
import TransformEmailForm from './Sheetforms/TransformEmailForm';
import TransformFirstNameForm from './Sheetforms/TransformFirstNameForm';
import TransformFloatForm from './Sheetforms/TransformFloatForm';
import TransformFullNameForm from './Sheetforms/TransformFullNameForm';
import TransformIntForm from './Sheetforms/TransformIntForm';
import TransformIntPhoneForm from './Sheetforms/TransformIntPhoneForm';
import TransformLastNameForm from './Sheetforms/TransformLastNameForm';
import TransformPhoneForm from './Sheetforms/TransformPhoneForm';
import TransformStringForm from './Sheetforms/TransformStringForm';

interface Props {
  transformer: TransformerWithType | undefined;
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
          disabled={!transformer}
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
  transformer: UserDefinedTransformer,
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
    case 'generate_e164_number':
      return (
        <GenerateE164NumberForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'generate_float':
      return (
        <GenerateFloatForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'generate_gender':
      return (
        <GenerateGenderForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'generate_int':
      return (
        <GenerateIntForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'generate_string':
      return (
        <GenerateStringForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'generate_string_phone':
      return (
        <GenerateStringPhoneForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'generate_uuid':
      return (
        <GenerateUuidForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'tranform_e164_number':
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
    case 'transform_float':
      return (
        <TransformFloatForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'tranform_full_name':
      return (
        <TransformFullNameForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'transform_int':
      return (
        <TransformIntForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'transform_int_phone':
      return (
        <TransformIntPhoneForm
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
    case 'transform_phone':
      return (
        <TransformPhoneForm
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
    default:
      <div>No transformer component found</div>;
  }
  return (
    <div>
      {' '}
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

interface TransformerMetadata {
  name: string;
  description: string;
  type: string;
}

export function handleTransformerMetadata(
  value: string | undefined
): TransformerMetadata {
  const tEntries: Record<string, TransformerMetadata>[] = [
    {
      passthrough: {
        name: 'Passthrough',
        description:
          'Passes the input value through to the desination with no changes.',
        type: 'passthrough',
      },
    },
    {
      null: {
        name: 'Null',
        description: 'Inserts a <null> string instead of the source value.',
        type: 'null',
      },
    },
    {
      invalid: {
        name: 'Invalid',
        description: 'Invalid transformer.',
        type: 'null',
      },
    },
  ];

  const def = {
    default: {
      name: 'Passthrough',
      description: 'Passthrough',
      type: 'passthrough',
    },
  };

  if (!value) {
    return def.default;
  }
  const res = tEntries.find((item) => item[value]);

  return res ? res[value] : def.default;
}

// merge system transformers into custom tranformers and add in additional metadata fields for system transformers to fit into the custom transformers interface
export function MergeSystemAndCustomTransformers(
  system: Transformer[],
  custom: CustomTransformer[]
): CustomTransformer[] {
  let merged: CustomTransformer[] = [...custom];

  system.map((st) => {
    const cf = {
      config: {
        case: st.config?.config.case,
        value: st.config?.config.value,
      },
    };
    const newCt = new CustomTransformer({
      name: handleTransformerMetadata(st.value).name,
      description: handleTransformerMetadata(st.value).description,
      type: handleTransformerMetadata(st.value).type,
      source: st.value,
      config: cf as TransformerConfig,
    });

    merged.push(newCt);
  });

  return merged;
}

/**
 * Returns only transformers that generate data with 0 input
 */
export function filterDataTransformers(
  transformers: Transformer[]
): Transformer[] {
  return transformers.filter(
    (t) => t.value !== 'passthrough' && t.value.startsWith('generate_')
  );
}
