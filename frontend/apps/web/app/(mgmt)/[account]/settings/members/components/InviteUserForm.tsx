'use client';

import ButtonText from '@/components/ButtonText';
import { CopyButton } from '@/components/CopyButton';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { getErrorMessage } from '@/util/util';
import { InviteMembersForm } from '@/yup-validations/invite-members';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { inviteUserToTeamAccount } from '@neosync/sdk/connectquery';
import { DialogClose } from '@radix-ui/react-dialog';
import { PlusIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';

interface Props {
  accountId: string;
  onInvited(): void;
}
export default function InviteUserForm(props: Props): ReactElement {
  const { accountId, onInvited } = props;
  const [showNewInviteDialog, setShowNewinviteDialog] = useState(false);
  const [newInviteToken, setNewInviteToken] = useState('');
  const [openInviteCreated, setOpenInviteCreated] = useState(false);

  const form = useForm<InviteMembersForm>({
    resolver: yupResolver(InviteMembersForm),
    defaultValues: {
      email: '',
    },
  });
  const { mutateAsync } = useMutation(inviteUserToTeamAccount);

  async function onSubmit(values: InviteMembersForm): Promise<void> {
    try {
      const invite = await mutateAsync({
        accountId: accountId,
        email: values.email,
      });
      setShowNewinviteDialog(false);
      if (invite?.invite?.token) {
        setNewInviteToken(invite.invite.token);
        setOpenInviteCreated(true);
      }
      onInvited();
      toast.success('Successfuly created user invite!');
      form.reset();
    } catch (err) {
      console.error(err);
      toast.error('Unable to create user invite.', {
        description: getErrorMessage(err),
      });
    }
  }

  function onOpenChange(open: boolean): void {
    setShowNewinviteDialog(open);
    form.reset();
  }

  return (
    <div>
      <Dialog open={showNewInviteDialog} onOpenChange={onOpenChange}>
        <DialogTrigger asChild>
          <Button type="button">
            <ButtonText leftIcon={<PlusIcon />} text="Invite Member" />
          </Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add new member</DialogTitle>
            <DialogDescription>
              Invite members with their email.
            </DialogDescription>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormControl>
                      <Input
                        autoCapitalize="off"
                        data-1p-ignore // tells 1password extension to not autofill this field
                        type="email"
                        id="email"
                        placeholder="example@email.com"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setShowNewinviteDialog(false)}
                  type="button"
                >
                  Cancel
                </Button>
                <Button type="submit">Submit</Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
      <InviteCreatedDialog
        open={openInviteCreated}
        setOpen={setOpenInviteCreated}
        token={newInviteToken}
      />
    </div>
  );
}
interface InviteCreatedDialogProps {
  open: boolean;
  setOpen: (value: boolean) => void;
  token: string;
}

function InviteCreatedDialog(props: InviteCreatedDialogProps): ReactElement {
  const { open, setOpen, token } = props;
  const { data: systemAppData } = useGetSystemAppConfig();
  const link = buildInviteLink(systemAppData?.publicAppBaseUrl ?? '', token);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent>
        <DialogHeader>
          <h3 className="font-semibold leading-none tracking-tight">
            Invite created!
          </h3>
          <p className="text-sm text-muted-foreground">
            Copy the invite link below and send it to the user.
          </p>
        </DialogHeader>
        <div className="flex flex-row gap-2">
          <Input value={link} readOnly />
          <CopyButton
            buttonVariant="outline"
            textToCopy={link}
            onCopiedText="Success!"
            onHoverText="Copy the invite link"
          />
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button type="button">
              <ButtonText text="Close" />
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export function buildInviteLink(baseUrl: string, token: string): string {
  return `${baseUrl}/invite?token=${token}`;
}
