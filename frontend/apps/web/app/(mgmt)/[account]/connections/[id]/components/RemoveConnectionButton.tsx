import ButtonText from '@/components/ButtonText';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { deleteConnection } from '@neosync/sdk';
import { TrashIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';

interface Props {
  connectionId: string;
}

export default function RemoveConnectionButton(props: Props): ReactElement {
  const { connectionId } = props;
  const router = useRouter();
  const account = useAccount();
  const { toast } = useToast();
  const { mutateAsync } = useMutation(deleteConnection);

  async function onDelete(): Promise<void> {
    try {
      await mutateAsync({ id: connectionId });
      toast({
        title: 'Successfully removed connection!',
        variant: 'success',
      });
      router.push(`/${account.account?.name}/connections`);
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to remove connection',
        description: getErrorMessage(err),
        variant: 'destructive',
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
