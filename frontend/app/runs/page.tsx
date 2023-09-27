'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { useGetJobRuns } from '@/libs/hooks/useGetJobRuns';
import { ReactElement } from 'react';
import { getColumns } from './components/JobRunsTable/columns';
import { DataTable } from './components/JobRunsTable/data-table';

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
      <JobRunsTable />
    </OverviewContainer>
  );
}

interface TableProps {}

function JobRunsTable(props: TableProps): ReactElement {
  const {} = props;
  const account = useAccount();
  const { isLoading, data, mutate } = useGetJobRuns(account?.id ?? '');

  if (isLoading) {
    return <SkeletonTable />;
  }

  const runs = data?.jobRuns ?? [];

  const columns = getColumns({
    onDeleted() {
      mutate();
    },
  });

  return (
    <div>
      <DataTable columns={columns} data={runs} />
    </div>
  );
}
