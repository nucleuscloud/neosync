import AccountPageRedirect from '@/components/AccountPageRedirect';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { ReactElement } from 'react';

export default function Home(): ReactElement {
  return (
    <AccountPageRedirect>
      <OverviewContainer
        Header={<PageHeader header={`Home`} />}
        containerClassName="home-page"
      >
        <div className="flex flex-col gap-4">
          <SkeletonTable />
        </div>
      </OverviewContainer>
    </AccountPageRedirect>
  );
}
