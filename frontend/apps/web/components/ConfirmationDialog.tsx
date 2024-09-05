import { AlertDialogTitle } from '@radix-ui/react-alert-dialog';
import { ReactElement, ReactNode, useState } from 'react';
import ButtonText from './ButtonText';
import Spinner from './Spinner';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTrigger,
} from './ui/alert-dialog';
import { Button, ButtonProps } from './ui/button';

interface Props {
  trigger: ReactNode;
  headerText: string;
  description: string;
  onConfirm(): void | Promise<void>;
  buttonText?: string;
  buttonVariant?: ButtonProps['variant'] | null | undefined;
  buttonIcon?: ReactNode;
}

export default function ConfirmationDialog(props: Props): ReactElement {
  const {
    trigger,
    headerText,
    description,
    onConfirm,
    buttonText = 'Confirm',
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
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>{trigger}</AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader className="gap-2">
          <AlertDialogTitle className="text-xl">{headerText}</AlertDialogTitle>
          <AlertDialogDescription className="tracking-tight">
            {description}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter className="w-full flex sm:justify-between pt-4">
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            asChild
            onClick={(e) => {
              e.preventDefault();
              onClick();
            }}
          >
            <Button type="button" variant={buttonVariant}>
              <ButtonText
                leftIcon={isTrying ? <Spinner /> : buttonIcon}
                text={buttonText}
              />
            </Button>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
