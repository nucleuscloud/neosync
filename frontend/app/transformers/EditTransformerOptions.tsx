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
import { ReactElement, useState } from 'react';
import EmailTransformerForm from './EmailTransformerForm';

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
      <SheetContent className="w-[800px]">
        <SheetHeader>
          <div className="flex flex-row justify-between w-full">
            <div className="flex flex-col space-y-2">
              <SheetTitle>
                {
                  handleTransformerMetadata(transformer ?? new Transformer())
                    .name
                }
              </SheetTitle>
              <SheetDescription>
                {
                  handleTransformerMetadata(transformer ?? new Transformer())
                    .description
                }
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
    default:
      <div>No transformer component found</div>;
  }
  return <div></div>;
}

const handleTransformerMetadata = (t: Transformer): TransformerMetadata => {
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
    { uuidv4: { name: 'UUIDv4', description: 'Generates a new UUIDv4 id.' } },
    {
      passthrough: {
        name: 'Passthrough',
        description:
          'Passes the input value through to the desination with no changes.',
      },
    },
  ];
  const res = tEntries.find((item) => item[t.value]);

  const def = {
    default: {
      name: 'Undefined',
      description: 'Undefined Transformer',
    },
  };

  return res ? res[t.value] : def.default;
};
