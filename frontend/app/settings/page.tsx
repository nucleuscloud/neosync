'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import MemberManagementSettings from './components/MemberManagementSettings';
import SubNav, { ITEMS } from './temporal/components/SubNav';

export default function Settings() {
  const { account } = useAccount();
  const isTeamAccount = account?.type.toString() == 'USER_ACCOUNT_TYPE_TEAM';
  return (
    <OverviewContainer
      Header={<PageHeader header="Settings" />}
      containerClassName="settings-page"
    >
      <div className="flex flex-col gap-4">
        <div>
          <SubNav items={ITEMS} />
        </div>
        {isTeamAccount && (
          <MemberManagementSettings accountId={account?.id || ''} />
        )}
      </div>
    </OverviewContainer>
  );
}
