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
  CustomTransformer,
  Transformer,
  TransformerConfig,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
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
  transformer: CustomTransformer | undefined;
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
                <Badge variant="outline">{transformer?.type}</Badge>
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
  transformer: CustomTransformer,
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
      generate_email: {
        name: 'Generate Email',
        description: 'Generates a new Generate email address.',
        type: 'string',
      },
    },
    {
      generate_realistic_email: {
        name: 'Generate Realistic Email',
        description: 'Generates a new realistic email address.',
        type: 'string',
      },
    },
    {
      transform_email: {
        name: 'Transform Email',
        description: 'Transforms an existing email address.',
        type: 'string',
      },
    },
    {
      generate_bool: {
        name: 'Generate Boolean',
        description: 'Generates a boolean value at random.',
        type: 'bool',
      },
    },
    {
      generate_card_number: {
        name: 'Generate Card Number',
        description: 'Generates a card number.',
        type: 'bool',
      },
    },
    {
      generate_city: {
        name: 'Generate City',
        description:
          'Randomly selects a city from a list of predefined US cities.',
        type: 'string',
      },
    },
    {
      generate_e164_number: {
        name: 'Generate E164 Phone Number',
        description: 'Generates a Generate phone number in e164 format.',
        type: 'string',
      },
    },
    {
      generate_first_name: {
        name: 'Generate First Name',
        description: 'Generates a Generate first name.',
        type: 'string',
      },
    },
    {
      generate_float: {
        name: 'Generate Float',
        description: 'Generates a Generate float value.',
        type: 'string',
      },
    },
    {
      generate_full_address: {
        name: 'Generate Full Address',
        description:
          'Randomly generates a street address in the format: {street_num} {street_addresss} {street_descriptor} {city}, {state} {zipcode}. For example, 123 Main Street Boston, Massachusetts 02169.',
        type: 'string',
      },
    },
    {
      generate_full_name: {
        name: 'Generate Full Name',
        description:
          'Generates a new full name consisting of a first and last name.',
        type: 'string',
      },
    },
    {
      generate_gender: {
        name: 'Gender',
        description:
          'Randomly generates one of the following genders: female, male, undefined, nonbinary.',
        type: 'string',
      },
    },
    {
      generate_int64_phone: {
        name: 'Generate Int64 Phone Number',
        description:
          'Generates a new phone number of type int64 with a default length of 10.',
        type: 'int64',
      },
    },
    {
      generate_int: {
        name: 'Generate Integer',
        description:
          'Generates a random integer value with a default length of 4 unless the Integer Length or Preserve Length paramters are defined. .',
        type: 'int64',
      },
    },
    {
      generate_last_name: {
        name: 'Generate Last Name',
        description: 'Generates a new last name.',
        type: 'string',
      },
    },
    {
      generate_sha256hash: {
        name: 'SHA256 Hash',
        description: 'SHA256 hashes a randomly generated value.',
        type: 'string',
      },
    },
    {
      generate_ssn: {
        name: 'Social Security Number',
        description:
          'Generates a completely random social security numbers including the hyphens in the format <xxx-xx-xxxx>',
        type: 'string',
      },
    },
    {
      generate_state: {
        name: 'State',
        description:
          'Randomly selects a US state and returns the two-character state code.',
        type: 'string',
      },
    },
    {
      generate_street_address: {
        name: 'Street Address',
        description:
          'Randomly generates a street address in the format: {street_num} {street_addresss} {street_descriptor}. For example, 123 Main Street.',
        type: 'string',
      },
    },
    {
      generate_string_phone: {
        name: 'String Phone',
        description:
          'Generates a Generate phone number and returns it as a string',
        type: 'string',
      },
    },
    {
      generate_string: {
        name: 'Generate String',
        description:
          'Creates a randomly ordered alphanumeric string with a default length of 10 unless the String Length parameter are defined.',
        type: 'string',
      },
    },
    {
      generate_unixtimestamp: {
        name: 'Unix Timestamp',
        description: 'Randomly generates a Unix timestamp.',
        type: 'int64',
      },
    },
    {
      generate_username: {
        name: 'Username',
        description:
          'Randomly generates a username in the format<first_initial><last_name>.',
        type: 'int64',
      },
    },
    {
      generate_utctimestamp: {
        name: 'UTC Timestamp',
        description: 'Randomly generates a UTC timestamp.',
        type: 'time',
      },
    },
    {
      generate_uuid: {
        name: 'UUID',
        description: 'Generates a new UUIDv4 id.',
        type: 'uuid',
      },
    },
    {
      generate_zipcode: {
        name: 'Zip Code',
        description:
          'Randomly selects a zip code from a list of predefined US cities.',
        type: 'string',
      },
    },
    {
      transform_e164_phone: {
        name: 'Transform E164 Phone',
        description: 'Transforms an existing E164 formatted phone number.',
        type: 'string',
      },
    },
    {
      transform_first_name: {
        name: 'Transform First Name',
        description: 'Transforms an existing first name.',
        type: 'string',
      },
    },
    {
      transform_float: {
        name: 'Transform Float',
        description: 'Transforms an existing float value.',
        type: 'string',
      },
    },
    {
      transform_full_name: {
        name: 'Transform Full Name',
        description: 'Transforms an existing full name.',
        type: 'string',
      },
    },
    {
      transform_int_phone: {
        name: 'Transform Integer Phone Number',
        description:
          'Transforms an existing phone number that is typed as an integer.',
        type: 'string',
      },
    },
    {
      transform_int: {
        name: 'Transform Integer ',
        description: 'Transforms an existing integer value.',
        type: 'string',
      },
    },
    {
      transform_last_name: {
        name: 'Transform Last Name ',
        description: 'Transforms an existing last name.',
        type: 'string',
      },
    },
    {
      transform_phone: {
        name: 'Transform String Phone Number ',
        description:
          'Transforms an existing phone number that is typed as a string.',
        type: 'string',
      },
    },
    {
      transform_string: {
        name: 'Transform String',
        description: 'Transforms an existing string value.',
        type: 'string',
      },
    },
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

// merge system into custom and add in additional metadata fields for system transformers
// to fit into the custom transformers interface
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
