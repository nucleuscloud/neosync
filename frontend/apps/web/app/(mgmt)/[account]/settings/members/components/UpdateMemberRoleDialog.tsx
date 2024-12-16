import { useAccount } from '@/components/providers/account-provider';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { getErrorMessage } from '@/util/util';
import { UpdateMemberRoleFormValues } from '@/yup-validations/invite-members';
import { useMutation } from '@connectrpc/connect-query';
import { AccountUser } from '@neosync/sdk';
import { setUserRole } from '@neosync/sdk/connectquery';
import { ReactElement, ReactNode, useState } from 'react';
import { toast } from 'sonner';
import UpdateMemberRoleForm from './UpdateMemberRoleForm';

interface Props {
  member: Pick<AccountUser, 'id' | 'name' | 'role' | 'email'>;
  onUpdated(): void;
  dialogButton: ReactNode;
}

export default function UpdateMemberRoleDialog(props: Props): ReactElement {
  const { member, onUpdated, dialogButton } = props;
  const { mutateAsync: updateUserRole } = useMutation(setUserRole);
  const { account } = useAccount();
  const [open, setOpen] = useState(false);

  async function onUpdate(values: UpdateMemberRoleFormValues): Promise<void> {
    if (!account) {
      return;
    }
    try {
      await updateUserRole({
        userId: member.id,
        role: values.role,
        accountId: account.id,
      });
      toast.success('Successfully updated user role!');
      onUpdated();
      setOpen(false);
    } catch (err) {
      console.error(err);
      toast.error('Unable to update user role', {
        description: getErrorMessage(err),
      });
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{dialogButton}</DialogTrigger>
      <DialogContent
        onPointerDownOutside={(e) => e.preventDefault()}
        className="max-h-[85vh] overflow-y-auto"
      >
        <DialogHeader>
          <DialogTitle>
            Update role for {member.name} ({member.email})
          </DialogTitle>
          <DialogDescription>Change the role of the user.</DialogDescription>
        </DialogHeader>
        <UpdateMemberRoleForm
          key={member.id}
          member={member}
          onSubmit={onUpdate}
          onCancel={() => setOpen(false)}
        />
      </DialogContent>
    </Dialog>
  );
}
