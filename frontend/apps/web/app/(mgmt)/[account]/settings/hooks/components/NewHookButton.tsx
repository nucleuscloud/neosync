import ButtonText from '@/components/ButtonText';
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
import { AccountHookService, NewAccountHook } from '@neosync/sdk';
import { PlusIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { toast } from 'sonner';
import NewHookForm from './NewHookForm';

interface Props {
  accountId: string;
  onCreated(): void;
}

export default function NewHookButton(props: Props): ReactElement {
  const { accountId, onCreated } = props;
  const { mutateAsync: createHook } = useMutation(
    AccountHookService.method.createAccountHook
  );
  const [open, setOpen] = useState(false);

  async function onCreate(values: NewAccountHook): Promise<void> {
    try {
      await createHook({
        accountId,
        hook: {
          config: values.config,
          description: values.description,
          enabled: values.enabled,
          name: values.name,
        },
      });
      toast.success('Successfully created account hook!');
      onCreated();
      setOpen(false);
    } catch (err) {
      console.error(err);
      toast.error('Unable to create account hook', {
        description: getErrorMessage(err),
      });
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button type="button">
          <ButtonText leftIcon={<PlusIcon />} text="New Hook" />
        </Button>
      </DialogTrigger>
      <DialogContent
        onPointerDownOutside={(e) => e.preventDefault()}
        className="lg:max-w-4xl max-h-[85vh] overflow-y-auto"
      >
        <DialogHeader>
          <DialogTitle>Create new Account Hook</DialogTitle>
          <DialogDescription>
            Configure values for a new account hook
          </DialogDescription>
        </DialogHeader>
        <NewHookForm onSubmit={onCreate} />
      </DialogContent>
    </Dialog>
  );
}
