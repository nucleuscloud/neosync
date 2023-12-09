'use client';
import { useAccount } from '@/components/providers/account-provider';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ReactElement } from 'react';
import { InvitesTable } from './components/InviteTable';
import MembersTable from './components/MemberTable';

interface Props {}

export default function MemberManagementSettings(_: Props): ReactElement {
  const { account, isLoading } = useAccount();
  const accountId = account?.id || '';

  if (isLoading) {
    return <Skeleton className="w-full h-12" />;
  }

  const isTeamAccount = account?.type.toString() == 'USER_ACCOUNT_TYPE_TEAM';
  if (!isTeamAccount) {
    return <div></div>;
  }

  return (
    <div className="mt-10">
      <h1 className="text-xl font-bold tracking-tight">Member Managment</h1>
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
