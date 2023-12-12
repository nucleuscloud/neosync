import { DialogClose } from '@radix-ui/react-dialog';
import { ReactElement, ReactNode, useState } from 'react';
import ButtonText from './ButtonText';
import Spinner from './Spinner';
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
  buttonText: string;
  buttonVariant:
    | 'default'
    | 'destructive'
    | 'outline'
    | 'secondary'
    | 'ghost'
    | 'link'
    | null
    | undefined;
  buttonIcon: ReactNode;
}

export default function ConfirmationDialog(props: Props): ReactElement {
  const {
    trigger,
    headerText,
    description,
    onConfirm,
    buttonText,
    buttonVariant,
    buttonIcon,
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
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent>
        <DialogHeader>{headerText}</DialogHeader>
        <DialogDescription>{description}</DialogDescription>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="secondary">
              <ButtonText text="Close" />
            </Button>
          </DialogClose>
          <Button
            type="submit"
            variant={buttonVariant}
            onClick={() => onClick()}
          >
            <ButtonText
              leftIcon={isTrying ? <Spinner /> : buttonIcon}
              text={buttonText}
            />
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
