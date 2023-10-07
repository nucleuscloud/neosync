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
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { Pencil1Icon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import { handleTransformerForm } from '../../[name]/components/transformer-component';

interface Props {
  transformer: Transformer | undefined;
}
export default function TransformerEdit(props: Props): ReactElement {
  const { transformer } = props;

  console.log('transformer', transformer);

  return (
    <Sheet>
      <SheetTrigger asChild>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger>
              <Button
                variant="outline"
                size="sm"
                disabled={!transformer}
                className="ml-auto hidden h-[36px] lg:flex"
              >
                <Pencil1Icon />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              {!transformer ? 'Select a Transformer' : 'Edit Transformer'}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </SheetTrigger>
      <SheetContent className="w-[800px]">
        <SheetHeader>
          <SheetTitle>{transformer?.title}</SheetTitle>
          <SheetDescription>{transformer?.description}</SheetDescription>
          <Separator />
        </SheetHeader>
        <div className="pt-8">
          {transformer && handleTransformerForm(transformer!)}
        </div>
      </SheetContent>
    </Sheet>
  );
}
