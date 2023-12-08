'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { ReactElement } from 'react';
import SubNav, { getNavSettings } from './components/SubNav';
import TemporalConfigForm from './components/TemporalConfigForm';

export default function Temporal(): ReactElement {
  const { account } = useAccount();
  return (
    <OverviewContainer
      Header={<PageHeader header="Temporal Settings" />}
      containerClassName="temporal-settings-page"
    >
      <div className="flex flex-col gap-4">
        <div>
          <SubNav items={getNavSettings(account?.name ?? '')} />
        </div>
        <TemporalConfigForm />
      </div>
    </OverviewContainer>
  );
}
