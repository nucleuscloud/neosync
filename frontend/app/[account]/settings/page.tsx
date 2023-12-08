'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { useGetAuthEnabled } from '@/libs/hooks/useGetAuthEnabled';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function Settings() {
  const authEnabled = useGetAuthEnabled();
  const { account, isLoading: isAccountLoading } = useAccount();

  const router = useRouter();

  useEffect(() => {
    if (!authEnabled) {
      router.push('/personal/settings/temporal');
    } else {
      router.push(`/${account?.name}/settings/members`);
    }
    if (!isAccountLoading && account?.name) {
      router.push(`/${account?.name}/settings/temporal`);
    }
  }, [account?.id, isAccountLoading]);
  return (
    <OverviewContainer
      Header={<PageHeader header="Settings" />}
      containerClassName="settings-page"
    >
      <div></div>
    </OverviewContainer>
  );
}
