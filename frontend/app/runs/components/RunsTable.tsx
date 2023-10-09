import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { useGetJobRuns } from '@/libs/hooks/useGetJobRuns';
import { ReactElement } from 'react';
import { getColumns } from './JobRunsTable/columns';
import { DataTable } from './JobRunsTable/data-table';

interface RunsTableProps {}

export default function RunsTable(props: RunsTableProps): ReactElement {
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
