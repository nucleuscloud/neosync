'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';
import RunsTable from './components/RunsTable';

export default function JobRuns(): ReactElement<any> {
  return (
    <OverviewContainer
      Header={<PageHeader header="Runs" />}
      containerClassName="runs-page"
    >
      <RunsTable />
    </OverviewContainer>
  );
}
