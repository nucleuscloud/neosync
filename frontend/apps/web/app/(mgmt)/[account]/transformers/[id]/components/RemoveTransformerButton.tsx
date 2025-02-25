import ButtonText from '@/components/ButtonText';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { TransformersService } from '@neosync/sdk';
import { TrashIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { toast } from 'sonner';

interface Props {
  transformerID: string;
}

export default function RemoveTransformerButton(props: Props): ReactElement<any> {
  const { transformerID } = props;
  const router = useRouter();
  const { account } = useAccount();
  const { mutateAsync } = useMutation(
    TransformersService.method.deleteUserDefinedTransformer
  );

  async function deleteTransformer(): Promise<void> {
    try {
      await mutateAsync({ transformerId: transformerID });
      toast.success('Successfully removed transformer!');
      router.push(`/${account?.name}/transformers`);
    } catch (err) {
      console.error(err);
      toast.error('Unable to remove transformer', {
        description: getErrorMessage(err),
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
