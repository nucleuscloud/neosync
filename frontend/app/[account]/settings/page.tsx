'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { useGetAuthEnabled } from '@/libs/hooks/useGetAuthEnabled';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import MemberManagementSettings from './components/MemberManagementSettings';
import SubNav, { ITEMS } from './temporal/components/SubNav';

export default function Settings() {
  const authEnabled = useGetAuthEnabled();
  const { account, isLoading: isAccountLoading } = useAccount();

  const router = useRouter();

  useEffect(() => {
    if (!authEnabled) {
      router.push('/personal/settings/temporal');
    }
    if (!isAccountLoading && account?.name) {
      router.push(`/${account.name}/settings/temporal`);
    }
  }, [account?.id, isAccountLoading]);
  return (
    <OverviewContainer
      Header={<PageHeader header="Settings" />}
      containerClassName="settings-page"
    >
      <div className="flex flex-row gap-4">
        <SubNav items={ITEMS} />
        <div>{authEnabled && <MemberManagementSettings />}</div>
      </div>
    </OverviewContainer>
  );
}
