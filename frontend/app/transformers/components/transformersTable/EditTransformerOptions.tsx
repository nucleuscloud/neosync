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
import { Pencil1Icon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import EmailTransformerForm from '../../[name]/components/EmailTransformerForm';
import FirstNameTransformerForm from '../../[name]/components/FirstNameForm';

interface Props {
  transformer: Transformer | undefined;
  index: number;
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
          <SheetTitle>{transformer?.title}</SheetTitle>
          <SheetDescription>{transformer?.description}</SheetDescription>
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

export function handleTransformerForm(
  transformer: Transformer,
  index?: number,
  setIsSheetOpen?: (val: boolean) => void
): ReactElement {
  switch (transformer.title) {
    case 'Email':
      return (
        <EmailTransformerForm index={index} setIsSheetOpen={setIsSheetOpen} />
      );
    case 'First Name':
      return <FirstNameTransformerForm transformer={transformer} />;
    default:
      <div>No transformer component found</div>;
  }
  return <div></div>;
}
