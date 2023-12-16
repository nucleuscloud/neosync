'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { UserAccountType } from '@neosync/sdk';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function Settings() {
  const { data: systemAppConfigData, isLoading: isSystemAppConfigDataLoading } =
    useGetSystemAppConfig();
  const { account, isLoading: isAccountLoading } = useAccount();

  const router = useRouter();

  useEffect(() => {
    if (isSystemAppConfigDataLoading || isAccountLoading) {
      return;
    }
    if (
      systemAppConfigData?.isAuthEnabled &&
      account?.name &&
      account.type === UserAccountType.TEAM
    ) {
      return router.push(`/${account?.name}/settings/members`);
    } else {
      return router.push('/personal/settings/temporal');
    }
  }, [
    account?.id,
    isAccountLoading,
    systemAppConfigData?.isAuthEnabled,
    isSystemAppConfigDataLoading,
  ]);
  return (
    <OverviewContainer
      Header={<PageHeader header="Settings" />}
      containerClassName="settings-page"
    >
      <div></div>
    </OverviewContainer>
  );
}
