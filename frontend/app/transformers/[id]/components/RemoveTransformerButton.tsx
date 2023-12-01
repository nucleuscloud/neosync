import ButtonText from '@/components/ButtonText';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
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

  async function deleteTransformer(): Promise<void> {
    try {
      await removeTransformer(transformerID);
      toast({
        title: 'Successfully removed transformer!',
        variant: 'success',
      });
      router.push(`/transformers`);
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

async function removeTransformer(transformerId: string): Promise<void> {
  const res = await fetch(
    `/api/transformers/custom?transformerId=${transformerId}`,
    {
      method: 'DELETE',
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}
