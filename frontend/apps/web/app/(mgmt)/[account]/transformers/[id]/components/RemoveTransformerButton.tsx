import ButtonText from '@/components/ButtonText';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { deleteUserDefinedTransformer } from '@neosync/sdk';
import { TrashIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';

interface Props {
  transformerID: string;
}

export default function RemoveTransformerButton(props: Props): ReactElement {
  const { transformerID } = props;
  const router = useRouter();
  const { toast } = useToast();
  const { account } = useAccount();
  const { mutateAsync } = useMutation(deleteUserDefinedTransformer);

  async function deleteTransformer(): Promise<void> {
    try {
      await mutateAsync({ transformerId: transformerID });
      toast({
        title: 'Successfully removed transformer!',
        variant: 'success',
      });
      router.push(`/${account?.name}/transformers`);
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to remove transformer',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }
  return (
    <DeleteConfirmationDialog
      trigger={
        <Button variant="destructive">
          <ButtonText leftIcon={<TrashIcon />} text="Delete Transformer" />
        </Button>
      }
      headerText="Are you sure you want to delete this Transformer?"
      description="Deleting this Transformer may impact Jobs that rely on it."
      onConfirm={async () => deleteTransformer()}
    />
  );
}
