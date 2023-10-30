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
import { CustomTransformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { Cross2Icon, Pencil1Icon } from '@radix-ui/react-icons';
import { ReactElement, useEffect, useRef, useState } from 'react';
import EmailTransformerForm from './Sheetforms/EmailTransformerForm';
import FirstNameTransformerForm from './Sheetforms/FirstnameTransformerForm';
import FullNameTransformerForm from './Sheetforms/FullnameTransformerForm';
import GenderTransformerForm from './Sheetforms/GenderTransformerForm';
import IntPhoneNumberTransformerForm from './Sheetforms/IntPhoneNumberTransformerForm';
import LastNameTransformerForm from './Sheetforms/LastnameTransformerForm';
import PhoneNumberTransformerForm from './Sheetforms/PhoneNumberTransformerForm';
import RandomIntTransformerForm from './Sheetforms/RandomIntTransformerForm';
import RandomStringTransformerForm from './Sheetforms/RandomStringTransformerForm';
import UuidTransformerForm from './Sheetforms/UuidTransformerForm';

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
    case 'email':
      return (
        <EmailTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'uuid':
      return (
        <UuidTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'first_name':
      return (
        <FirstNameTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'last_name':
      return (
        <LastNameTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'full_name':
      return (
        <FullNameTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'phone_number':
      return (
        <PhoneNumberTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'int_phone_number':
      return (
        <IntPhoneNumberTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'random_string':
      return (
        <RandomStringTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    case 'random_int':
      return (
        <RandomIntTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    // case 'random_float':
    //   return (
    //     <RandomFloatTransformerForm
    //       index={index}
    //       setIsSheetOpen={setIsSheetOpen}
    //     />
    //   );
    case 'gender':
      return (
        <GenderTransformerForm
          index={index}
          setIsSheetOpen={setIsSheetOpen}
          transformer={transformer}
        />
      );
    default:
      <div>No transformer component found</div>;
  }
  return <div></div>;
}
