import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { useGetJobRunEvents } from '@/libs/hooks/useGetJobRunEvents';
import { ReactElement } from 'react';
import { getColumns } from './JobRunActivityTable/columns';
import { DataTable } from './JobRunActivityTable/data-table';

interface JobRunActivityTableProps {
  jobId: string;
}

export default function JobRunActivityTable(
  props: JobRunActivityTableProps
): ReactElement {
  const { jobId } = props;
  const { data, isLoading } = useGetJobRunEvents(jobId);

  const events = data?.events || [];

  if (isLoading) {
    return <SkeletonTable />;
  }

  const columns = getColumns({});
  const isError = events.some((e) => e.tasks.some((t) => t.error));

  return (
    <div>
      <DataTable columns={columns} data={events} isError={isError} />
    </div>
  );
}
