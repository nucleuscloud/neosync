'use client';
import ButtonText from '@/components/ButtonText';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { deleteAccountApiKey } from '@neosync/sdk';
import { TrashIcon } from '@radix-ui/react-icons';
import { ReactElement, ReactNode } from 'react';

interface Props {
  id: string;
  trigger?: ReactNode;
  onDeleted?(): void;
}

export default function RemoveAccountApiKeyButton(props: Props): ReactElement {
  const { id, trigger, onDeleted } = props;
  const { toast } = useToast();
  const { mutateAsync } = useMutation(deleteAccountApiKey);

  async function onDelete(): Promise<void> {
    try {
      await mutateAsync({ id: id });
      toast({
        title: 'Successfully removed api key!',
        variant: 'success',
      });
      if (onDeleted) {
        onDeleted();
      }
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to remove api key!',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <DeleteConfirmationDialog
      trigger={
        trigger ? (
          trigger
        ) : (
          <Button variant="destructive">
            <ButtonText leftIcon={<TrashIcon />} text="Delete API Key" />
          </Button>
        )
      }
      headerText="Are you sure you want to delete this api key?"
      description="Deleting this api key is irreversable and may break running workloads!"
      onConfirm={async () => onDelete()}
    />
  );
}
