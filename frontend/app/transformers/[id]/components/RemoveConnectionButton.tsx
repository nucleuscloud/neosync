import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
import { TrashIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';

interface Props {
  connectionId: string;
}

export default function RemoveConnectionButton(props: Props): ReactElement {
  const { connectionId } = props;
  const router = useRouter();
  const { toast } = useToast();
  return (
    <Button
      variant="destructive"
      onClick={async () => {
        try {
          await removeConnection(connectionId);
          toast({
            title: 'Successfully removed connection!',
            variant: 'default',
          });
          router.push(`/connections`);
        } catch (err) {
          console.error(err);
          toast({
            title: 'Unable to remove connection',
            description: getErrorMessage(err),
            variant: 'destructive',
          });
        }
      }}
    >
      <div className="flex flex-row gap-1 items-center">
        <TrashIcon />
        <p>Delete Connection</p>
      </div>
    </Button>
  );
}

async function removeConnection(connectionId: string): Promise<void> {
  const res = await fetch(`/api/connections/${connectionId}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}
