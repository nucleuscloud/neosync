import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';
import SubNav, { ITEMS } from '../temporal/components/SubNav';

export default function ApiKeys(): ReactElement {
  return (
    <OverviewContainer
      Header={<PageHeader header="API Keys" />}
      containerClassName="apikeys-settings-page"
    >
      <div className="flex flex-col gap-4">
        <div>
          <SubNav items={ITEMS} />
        </div>
        <div>Todo</div>
      </div>
    </OverviewContainer>
  );
}
