import { DialogClose } from '@radix-ui/react-dialog';
import { TrashIcon } from '@radix-ui/react-icons';
import { ReactElement, ReactNode, useState } from 'react';
import ButtonText from './ButtonText';
import { Button } from './ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTrigger,
} from './ui/dialog';

interface Props {
  trigger: ReactNode;
  headerText: string;
  description: string;
  onConfirm(): void | Promise<void>;
  deleteButtonText?: string;
}

export default function DeleteConfirmationDialog(props: Props): ReactElement {
  const {
    trigger,
    headerText,
    description,
    onConfirm,
    deleteButtonText = 'Delete',
  } = props;
  const [open, setOpen] = useState(false);
  const [isTrying, setIsTrying] = useState(false);

  async function onClick(): Promise<void> {
    if (isTrying) {
      return;
    }
    setIsTrying(true);
    try {
      await onConfirm();
      setOpen(false);
    } finally {
      setIsTrying(false);
    }
  }
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger>{trigger}</DialogTrigger>
      <DialogContent>
        <DialogHeader>{headerText}</DialogHeader>
        <DialogDescription>{description}</DialogDescription>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="secondary">
              <ButtonText text="Close" />
            </Button>
          </DialogClose>
          <Button type="submit" variant="destructive" onClick={() => onClick()}>
            <ButtonText leftIcon={<TrashIcon />} text={deleteButtonText} />
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
