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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { PlusIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { InvitesTable } from './InviteTable';
import { MembersTable } from './MemberTable';

export default function MemberManagementSettings(): ReactElement {
  const { account } = useAccount();
  const [showNewInviteDialog, setShowNewinviteDialog] = useState(false);

  if (account?.type.toString() != 'USER_ACCOUNT_TYPE_TEAM') {
    return <></>;
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
            <MembersTable accountId={account?.id || ''} />
          </TabsContent>
          <TabsContent value="invites">
            <InvitesTable accountId={account?.id || ''} />
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
                  id="email"
                  placeholder="example@email.com"
                  onChange={(event) => setTeamName(event.target.value)}
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
            <Button type="submit" onClick={() => onSubmit(teamName)}>
              Continue
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
