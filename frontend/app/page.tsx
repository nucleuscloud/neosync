'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import RunsTable from './runs/components/RunsTable';

export default function Home() {
  return (
    <div className="mx-24">
      <OverviewContainer
        Header={<PageHeader header="Latest Job Runs" />}
        containerClassName="overview-page"
      >
        <RunsTable />
      </OverviewContainer>
    </div>
  );
}
