import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { Button } from '@/components/ui/button';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { JobHook } from '@neosync/sdk';
import { deleteJobHook } from '@neosync/sdk/connectquery';
import { TrashIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import { toast } from 'sonner';

interface Props {
  onDeleted(): void;
  hook: Pick<JobHook, 'id' | 'name'>;
}
export default function RemoveHookButton(props: Props): ReactElement {
  const { hook, onDeleted } = props;
  const { mutateAsync: removeHook } = useMutation(deleteJobHook);

  async function onDelete(): Promise<void> {
    try {
      await removeHook({ id: hook.id });
      toast.success('Successfully removed job hook!');
      onDeleted();
    } catch (err) {
      console.error(err);
      toast.error('Unable to remove job hook', {
        description: getErrorMessage(err),
      });
    }
  }

  return (
    <DeleteConfirmationDialog
      trigger={
        <Button variant="destructive" type="button">
          <TrashIcon />
        </Button>
      }
      headerText={`Are you sure you want to delete job hook: ${hook.name}?`}
      description="Deleting this hook is irreversable!"
      onConfirm={async () => onDelete()}
    />
  );
}
