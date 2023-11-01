import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';
import TemporalConfigForm from './components/TemporalConfigForm';

export default function Temporal(): ReactElement {
  return (
    <OverviewContainer
      Header={<PageHeader header="Temporal Settings" />}
      containerClassName="temporal-settings-page"
    >
      <div>
        <TemporalConfigForm />
      </div>
    </OverviewContainer>
  );
}
