'use client';
import ButtonText from '@/components/ButtonText';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { Button } from '@/components/ui/button';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { ApiKeyService } from '@neosync/sdk';
import { TrashIcon } from '@radix-ui/react-icons';
import { ReactElement, ReactNode } from 'react';
import { toast } from 'sonner';

interface Props {
  id: string;
  trigger?: ReactNode;
  onDeleted?(): void;
}

export default function RemoveAccountApiKeyButton(props: Props): ReactElement {
  const { id, trigger, onDeleted } = props;
  const { mutateAsync } = useMutation(ApiKeyService.method.deleteAccountApiKey);

  async function onDelete(): Promise<void> {
    try {
      await mutateAsync({ id: id });
      toast.success('Successfully removed api key!');
      if (onDeleted) {
        onDeleted();
      }
    } catch (err) {
      console.error(err);
      toast.error('Unable to remove api key!', {
        description: getErrorMessage(err),
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
