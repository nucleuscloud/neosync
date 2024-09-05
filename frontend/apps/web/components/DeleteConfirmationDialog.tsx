import { TrashIcon } from '@radix-ui/react-icons';
import { ReactElement, ReactNode } from 'react';
import ConfirmationDialog from './ConfirmationDialog';

interface Props {
  trigger: ReactNode;
  headerText: string;
  description: string;
  onConfirm(): void | Promise<void>;
  deleteButtonText?: string;
}

export default function DeleteConfirmationDialog(props: Props): ReactElement {
  const { deleteButtonText = 'Delete' } = props;

  return (
    <ConfirmationDialog
      {...props}
      buttonText={deleteButtonText}
      buttonIcon={<TrashIcon />}
      buttonVariant="destructive"
    />
  );
}
