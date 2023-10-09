'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';
import RunsTable from './components/RunsTable';

export default function JobRuns(): ReactElement {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Runs"
          description="Create and manage job runs to send and receive data"
        />
      }
      containerClassName="runs-page"
    >
      <RunsTable />
    </OverviewContainer>
  );
}
