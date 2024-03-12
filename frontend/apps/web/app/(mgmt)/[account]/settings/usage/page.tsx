import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';

export default function UsagePage(): ReactElement {
  return (
    <OverviewContainer
      Header={<PageHeader header="Usage" />}
      containerClassName="usage-page"
    >
      <div>Hello world</div>
    </OverviewContainer>
  );
}
