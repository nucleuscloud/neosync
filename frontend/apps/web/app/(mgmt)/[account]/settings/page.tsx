'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect } from 'react';

export default function Settings(): ReactElement<any> {
  const { account, isLoading: isAccountLoading } = useAccount();
  const router = useRouter();
  useEffect(() => {
    if (isAccountLoading) {
      return;
    }
    const accountName = account?.name ?? 'personal';
    return router.push(`/${accountName}/settings/api-keys`);
  }, [account?.name, isAccountLoading]);

  return (
    <OverviewContainer
      Header={<PageHeader header="Settings" />}
      containerClassName="settings-page"
    >
      <div />
    </OverviewContainer>
  );
}
