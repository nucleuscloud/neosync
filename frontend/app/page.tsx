'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import RunsTable from './runs/components/RunsTable';

export default function Home() {
  return (
    <OverviewContainer
      Header={<PageHeader header="Latest Job Runs" description="" />}
      containerClassName="overview-page"
    >
      <RunsTable />
    </OverviewContainer>
  );
}
