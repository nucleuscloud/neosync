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
import LastNameTransformerForm from './forms/LastnameTransformerForm';
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

  return (
    <Sheet open={isSheetOpen} onOpenChange={() => setIsSheetOpen(true)}>
      <SheetTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          disabled={!transformer || transformer.value == 'passthrough'}
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
    default:
      <div>No transformer component found</div>;
  }
  return <div></div>;
}

function handleTransformerMetadata(
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
        description: 'Anonymizes or generates a new phone number.',
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
