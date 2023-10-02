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
  return (
    <Button
      variant="destructive"
      onClick={async () => {
        try {
          await removeTransformer(transformerID);
          toast({
            title: 'Successfully removed transformer!',
            variant: 'default',
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
      }}
    >
      <div className="flex flex-row gap-1 items-center">
        <TrashIcon />
        <p>Delete Transformer</p>
      </div>
    </Button>
  );
}

async function removeTransformer(transformerID: string): Promise<void> {
  const res = await fetch(`/api/transformer/${transformerID}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}
