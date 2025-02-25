import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { Button } from '@/components/ui/button';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { AccountHook, AccountHookService } from '@neosync/sdk';
import { TrashIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import { toast } from 'sonner';

interface Props {
  onDeleted(): void;
  hook: Pick<AccountHook, 'id' | 'name'>;
}
export default function RemoveHookButton(props: Props): ReactElement<any> {
  const { hook, onDeleted } = props;
  const { mutateAsync: removeHook } = useMutation(
    AccountHookService.method.deleteAccountHook
  );

  async function onDelete(): Promise<void> {
    try {
      await removeHook({ id: hook.id });
      toast.success('Successfully removed account hook!');
      onDeleted();
    } catch (err) {
      console.error(err);
      toast.error('Unable to remove account hook', {
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
      headerText={`Are you sure you want to delete account hook: ${hook.name}?`}
      description="Deleting this hook is irreversable!"
      onConfirm={async () => onDelete()}
    />
  );
}
