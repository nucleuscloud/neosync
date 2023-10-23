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
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { Cross2Icon, Pencil1Icon } from '@radix-ui/react-icons';
import { ReactElement, useEffect, useRef, useState } from 'react';
import EmailTransformerForm from './forms/EmailTransformerForm';
import FirstNameTransformerForm from './forms/FirstnameTransformerForm';
import FullNameTransformerForm from './forms/FullnameTransformerForm';
import GenderTransformerForm from './forms/GenderTransformerForm';
import IntPhoneNumberTransformerForm from './forms/IntPhoneNumberTransformerForm';
import LastNameTransformerForm from './forms/LastnameTransformerForm';
import PhoneNumberTransformerForm from './forms/PhoneNumberTransformerForm';
import RandomFloatTransformerForm from './forms/RandomFloatTransformerForm';
import RandomIntTransformerForm from './forms/RandomIntTransformerForm';
import RandomStringTransformerForm from './forms/RandomStringTransformerForm';
import UuidTransformerForm from './forms/UuidTransformerForm';

interface Props {
  transformer: Transformer | undefined;
  index: number;
}

interface TransformerMetadata {
  name: string;
  description: string;
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

  const disabledSheetValues = ['passthrough', 'null']; // the sheet button will be disabled for any transformer with these values

  const handleDisableSheet = () => {
    if (!transformer) {
      return true;
    }

    return disabledSheetValues.includes(transformer.value);
  };

  return (
    <Sheet open={isSheetOpen} onOpenChange={() => setIsSheetOpen(true)}>
      <SheetTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          disabled={handleDisableSheet()}
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
              <SheetTitle>
                {handleTransformerMetadata(transformer).name}
              </SheetTitle>
              <SheetDescription>
                {handleTransformerMetadata(transformer).description}
              </SheetDescription>
            </div>
            <Button variant="ghost" onClick={() => setIsSheetOpen(false)}>
              <Cross2Icon className="h-4 w-4" />
            </Button>
          </div>
          <Separator />
        </SheetHeader>
        <div className="pt-8">
          {transformer &&
            handleTransformerForm(transformer!, index, setIsSheetOpen)}
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
  switch (transformer.value) {
    case 'email':
      return (
        <EmailTransformerForm index={index} setIsSheetOpen={setIsSheetOpen} />
      );
    case 'uuid':
      return (
        <UuidTransformerForm index={index} setIsSheetOpen={setIsSheetOpen} />
      );
    case 'first_name':
      return (
        <FirstNameTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
        />
      );
    case 'last_name':
      return (
        <LastNameTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
        />
      );
    case 'full_name':
      return (
        <FullNameTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
        />
      );
    case 'phone_number':
      return (
        <PhoneNumberTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
        />
      );
    case 'int_phone_number':
      return (
        <IntPhoneNumberTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
        />
      );
    case 'random_string':
      return (
        <RandomStringTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
        />
      );
    case 'random_int':
      return (
        <RandomIntTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
        />
      );
    case 'random_float':
      return (
        <RandomFloatTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
        />
      );
    case 'gender':
      return (
        <GenderTransformerForm index={index} setIsSheetOpen={setIsSheetOpen} />
      );
    default:
      <div>No transformer component found</div>;
  }
  return <div></div>;
}

export function handleTransformerMetadata(
  t: Transformer | undefined
): TransformerMetadata {
  const tEntries: Record<string, TransformerMetadata>[] = [
    {
      email: {
        name: 'Email',
        description: 'Anonymizes or generates a new email.',
      },
    },
    {
      phone_number: {
        name: 'Phone Number',
        description:
          'Anonymizes or generates a new phone number. The default format is <XXX-XXX-XXXX>.',
      },
    },
    {
      int_phone_number: {
        name: 'Int64 Phone Number',
        description:
          'Anonymizes or generates a new phone number of type int64 with a default length of 10.',
      },
    },
    {
      first_name: {
        name: 'First Name',
        description: 'Anonymizes or generates a new first name.',
      },
    },
    {
      last_name: {
        name: 'Last Name',
        description: 'Anonymizes or generates a new last name.',
      },
    },
    {
      full_name: {
        name: 'Full Name',
        description:
          'Anonymizes or generates a new full name consisting of a first and last name.',
      },
    },
    { uuid: { name: 'UUID', description: 'Generates a new UUIDv4 id.' } },
    {
      passthrough: {
        name: 'Passthrough',
        description:
          'Passes the input value through to the desination with no changes.',
      },
    },
    {
      null: {
        name: 'Null',
        description: 'Inserts a <null> string instead of the source value.',
      },
    },
    {
      random_string: {
        name: 'Random String',
        description:
          'Creates a randomly ordered alphanumeric string with a default length of 10 unless the String Length or Preserve Length parameters are defined.',
      },
    },
    {
      random_bool: {
        name: 'Random Bool',
        description: 'Generates a boolean value at random.',
      },
    },
    {
      random_int: {
        name: 'Random Integer',
        description:
          'Generates a random integer value with a default length of 4 unless the Integer Length or Preserve Length paramters are defined. .',
      },
    },
    {
      random_float: {
        name: 'Random Float',
        description:
          'Generates a random float value with a default length of <XX.XXX>.',
      },
    },
    {
      gender: {
        name: 'Gender',
        description:
          'Randomly generates one of the following genders: female, male, undefined, nonbinary.',
      },
    },
  ];

  const def = {
    default: {
      name: 'Undefined',
      description: 'Undefined Transformer',
    },
  };

  if (!t) {
    return def.default;
  }
  const res = tEntries.find((item) => item[t.value]);

  return res ? res[t.value] : def.default;
}
