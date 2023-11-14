'use client';
import ButtonText from '@/components/ButtonText';
import SubPageHeader from '@/components/headers/SubPageHeader';
import { useAccount } from '@/components/providers/account-provider';
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
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { toast } from '@/components/ui/use-toast';
import {
  InviteUserToTeamAccountRequest,
  InviteUserToTeamAccountResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { getErrorMessage } from '@/util/util';
import { DialogClose } from '@radix-ui/react-dialog';
import { PlusIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { InvitesTable } from './InviteTable';
import { MembersTable } from './MemberTable';

interface Props {}

export default function MemberManagementSettings(_: Props): ReactElement {
  const { account, isLoading } = useAccount();
  const accountId = account?.id || '';
  const [showNewInviteDialog, setShowNewinviteDialog] = useState(false);
  const [inviteEmail, setInviteEmail] = useState('');
  const [openInviteCreated, setOpenInviteCreated] = useState(false);
  const [newInviteToken, setNewInviteToken] = useState('');

  async function onSubmit(email: string): Promise<void> {
    try {
      const invite = await inviteUserToTeamAccount(accountId, email);
      setShowNewinviteDialog(false);
      if (invite?.invite?.token) {
        setNewInviteToken(invite.invite.token);
        setOpenInviteCreated(true);
      }
      toast({
        title: 'Successfully created invite!',
        variant: 'success',
      });
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to create invite',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  if (isLoading) {
    return <Skeleton className="w-full h-12" />;
  }

  const isTeamAccount = account?.type.toString() == 'USER_ACCOUNT_TYPE_TEAM';
  if (!isTeamAccount) {
    return <div></div>;
  }

  return (
    <div className="mt-10">
      <InviteCreatedDialog
        open={openInviteCreated}
        setOpen={setOpenInviteCreated}
        token={newInviteToken}
      />
      <Dialog open={showNewInviteDialog} onOpenChange={setShowNewinviteDialog}>
        <SubPageHeader
          header="Members and Invites"
          description="Manage members in your account, as well as invite new members."
          extraHeading={
            <DialogTrigger asChild>
              <Button>
                <ButtonText leftIcon={<PlusIcon />} text="Invite Member" />
              </Button>
            </DialogTrigger>
          }
        />
        <Tabs defaultValue="members">
          <TabsList>
            <TabsTrigger value="members">Members</TabsTrigger>
            <TabsTrigger value="invites">Invites</TabsTrigger>
          </TabsList>
          <TabsContent value="members">
            <MembersTable accountId={accountId} />
          </TabsContent>
          <TabsContent value="invites">
            <InvitesTable accountId={accountId} />
          </TabsContent>
        </Tabs>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add new member</DialogTitle>
            <DialogDescription>
              Invite members with their email.
            </DialogDescription>
          </DialogHeader>
          <div>
            <div className="space-y-4 py-2 pb-4">
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  type="email"
                  id="email"
                  placeholder="example@email.com"
                  onChange={(event) => setInviteEmail(event.target.value)}
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowNewinviteDialog(false)}
            >
              Cancel
            </Button>
            <Button type="submit" onClick={() => onSubmit(inviteEmail)}>
              Continue
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
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

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent>
        <DialogHeader>Invite created!</DialogHeader>
        <DialogDescription>{token}</DialogDescription>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="secondary">
              <ButtonText text="Close" />
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

async function inviteUserToTeamAccount(
  accountId: string,
  email: string
): Promise<InviteUserToTeamAccountResponse | undefined> {
  const res = await fetch(`/api/users/accounts/${accountId}/invites`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(
      new InviteUserToTeamAccountRequest({
        email,
        accountId,
      })
    ),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return InviteUserToTeamAccountResponse.fromJson(await res.json());
}
