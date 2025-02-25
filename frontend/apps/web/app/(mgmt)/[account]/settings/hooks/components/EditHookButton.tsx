import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { AccountHook, AccountHookService } from '@neosync/sdk';
import { Pencil1Icon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { toast } from 'sonner';
import { EditHookForm } from './EditHookForm';

interface Props {
  onEdited(): void;
  hook: AccountHook;
}

export default function EditHookButton(props: Props): ReactElement<any> {
  const { hook, onEdited } = props;
  const { mutateAsync: updateHook } = useMutation(
    AccountHookService.method.updateAccountHook
  );
  const [open, setOpen] = useState(false);

  async function onUpdate(values: Partial<AccountHook>): Promise<void> {
    try {
      await updateHook({
        id: hook.id,
        config: values.config,
        description: values.description,
        enabled: values.enabled,
        name: values.name,
        events: values.events,
      });
      toast.success('Successfully updated account hook!');
      onEdited();
      setOpen(false);
    } catch (err) {
      console.error(err);
      toast.error('Unable to update account hook', {
        description: getErrorMessage(err),
      });
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" type="button">
          <Pencil1Icon />
        </Button>
      </DialogTrigger>
      <DialogContent
        onPointerDownOutside={(e) => e.preventDefault()}
        className="lg:max-w-4xl max-h-[85vh] overflow-y-auto"
      >
        <DialogHeader>
          <DialogTitle>Edit Account Hook: {hook.name}</DialogTitle>
          <DialogDescription>
            Change any of the available account hook settings.
          </DialogDescription>
        </DialogHeader>
        <EditHookForm key={hook.id} hook={hook} onSubmit={onUpdate} />
      </DialogContent>
    </Dialog>
  );
}
