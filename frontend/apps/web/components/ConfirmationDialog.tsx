import { cn } from '@/libs/utils';
import { AlertDialogTitle } from '@radix-ui/react-alert-dialog';
import { ReactElement, ReactNode, useState, type JSX } from 'react';
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
import { Button, ButtonProps, buttonVariants } from './ui/button';

export interface Props {
  trigger?: ReactNode;
  headerText?: string;
  description?: string;
  body?: JSX.Element;
  buttonText?: string;
  buttonVariant?: ButtonProps['variant'] | null | undefined;
  buttonIcon?: ReactNode;
  containerClassName?: string;
  onConfirm(): void | Promise<void>;
}

export default function ConfirmationDialog(props: Props): ReactElement<any> {
  const {
    trigger = <Button type="button">Press to Confirm</Button>,
    headerText = 'Are you sure?',
    description = 'This will confirm the action that you selected.',
    body,
    buttonText = 'Confirm',
    buttonVariant,
    buttonIcon,
    containerClassName,
    onConfirm,
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
      <AlertDialogContent className={cn(containerClassName)}>
        <AlertDialogHeader className="gap-2">
          <AlertDialogTitle className="text-xl">{headerText}</AlertDialogTitle>
          <AlertDialogDescription className="tracking-tight">
            {description}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <div className="pt-6">{body}</div>
        <AlertDialogFooter className="w-full flex sm:justify-between mt-4">
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={(e) => {
              e.preventDefault();
              onClick();
            }}
            className={cn(buttonVariants({ variant: buttonVariant }))}
          >
            <ButtonText
              leftIcon={isTrying ? <Spinner /> : buttonIcon}
              text={buttonText}
            />
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
