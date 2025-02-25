import LearnMoreLink from '@/components/labels/LearnMoreLink';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Separator } from '@/components/ui/separator';
import { ReactElement } from 'react';

interface Props {
  open: boolean;
  onOpenChange(open: boolean): void;

  body: ReactElement<any>;
}

export default function EditItemDialog(props: Props): ReactElement<any> {
  const { open, onOpenChange, body } = props;
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="max-w-5xl"
        onPointerDownOutside={(e) => e.preventDefault()}
        onEscapeKeyDown={(e) => e.preventDefault()}
      >
        <DialogHeader>
          <div className="flex flex-row w-full">
            <div className="flex flex-col space-y-2 w-full">
              <div className="flex flex-row justify-between items-center">
                <div className="flex flex-row gap-4">
                  <DialogTitle className="text-xl">Subset Query</DialogTitle>
                </div>
              </div>
              <div className="flex flex-row items-center gap-2">
                <DialogDescription>
                  Subset your data using SQL expressions.{' '}
                  <LearnMoreLink href="https://docs.neosync.dev/table-constraints/subsetting" />
                </DialogDescription>
              </div>
            </div>
          </div>
          <Separator />
        </DialogHeader>
        <div className="pt-4">{body}</div>
      </DialogContent>
    </Dialog>
  );
}
