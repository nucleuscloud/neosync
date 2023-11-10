'use client';
import ButtonText from '@/components/ButtonText';
import SubPageHeader from '@/components/headers/SubPageHeader';
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { toast } from '@/components/ui/use-toast';
import {
  InviteUserToTeamAccountRequest,
  InviteUserToTeamAccountResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { getErrorMessage } from '@/util/util';
import { PlusIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { InvitesTable } from './InviteTable';
import { MembersTable } from './MemberTable';

interface Props {
  accountId: string;
}

export default function MemberManagementSettings(props: Props): ReactElement {
  const { accountId } = props;
  const [showNewInviteDialog, setShowNewinviteDialog] = useState(false);
  const [inviteEmail, setInviteEmail] = useState('');

  async function onSubmit(email: string): Promise<void> {
    try {
      await inviteUserToTeamAccount(accountId, email);
      setShowNewinviteDialog(false);
      toast({
        title: 'Successfully created team!',
        variant: 'success',
      });
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to create team',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <div className="mt-10">
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
