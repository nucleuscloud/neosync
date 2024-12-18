import ButtonText from '@/components/ButtonText';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { ConnectionService } from '@neosync/sdk';
import { TrashIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { toast } from 'sonner';

interface Props {
  connectionId: string;
}

export default function RemoveConnectionButton(props: Props): ReactElement {
  const { connectionId } = props;
  const router = useRouter();
  const account = useAccount();
  const { mutateAsync } = useMutation(
    ConnectionService.method.deleteConnection
  );

  async function onDelete(): Promise<void> {
    try {
      await mutateAsync({ id: connectionId });
      toast.success('Successfully removed connection!');
      router.push(`/${account.account?.name}/connections`);
    } catch (err) {
      console.error(err);
      toast.error('Unable to remove connection', {
        description: getErrorMessage(err),
      });
    }
  }

  return (
    <DeleteConfirmationDialog
      trigger={
        <Button variant="destructive">
          <ButtonText leftIcon={<TrashIcon />} text="Delete Connection" />
        </Button>
      }
      headerText="Are you sure you want to delete this connection?"
      description="Deleting this connection is irreversable!"
      onConfirm={async () => onDelete()}
    />
  );
}
