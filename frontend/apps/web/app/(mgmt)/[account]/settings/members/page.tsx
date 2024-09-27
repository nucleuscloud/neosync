'use client';
import SubPageHeader from '@/components/headers/SubPageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useQuery } from '@connectrpc/connect-query';
import { UserAccountType } from '@neosync/sdk';
import { getTeamAccountInvites } from '@neosync/sdk/connectquery';
import { ReactElement } from 'react';
import { InvitesTable } from './components/InviteTable';
import InviteUserForm from './components/InviteUserForm';
import MembersTable from './components/MemberTable';

interface Props {}

export default function MemberManagementSettings(_: Props): ReactElement {
  const { account, isLoading } = useAccount();
  const accountId = account?.id || '';

  const { refetch } = useQuery(
    getTeamAccountInvites,
    { accountId: accountId },
    { enabled: !!accountId }
  );

  if (isLoading) {
    return <Skeleton className="w-full h-12" />;
  }

  const isTeamAccount = account?.type === UserAccountType.TEAM;
  if (!isTeamAccount) {
    return (
      <div>
        <Alert variant="destructive">
          <AlertTitle>Members can only be added to team accounts</AlertTitle>
        </Alert>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-5">
      <SubPageHeader
        header="Member Management"
        description={`Manage your account's members and invites`}
        extraHeading={
          <InviteUserForm accountId={accountId} onInvited={() => refetch()} />
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
    </div>
  );
}
