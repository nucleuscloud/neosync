import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';
import SubNav, { ITEMS } from './components/SubNav';
import TemporalConfigForm from './components/TemporalConfigForm';

export default function Temporal(): ReactElement {
  return (
    <OverviewContainer
      Header={<PageHeader header="Temporal Settings" />}
      containerClassName="temporal-settings-page mx-24"
    >
      <div className="flex flex-col gap-4">
        <div>
          <SubNav items={ITEMS} />
        </div>
        <TemporalConfigForm />
      </div>
    </OverviewContainer>
  );
}
