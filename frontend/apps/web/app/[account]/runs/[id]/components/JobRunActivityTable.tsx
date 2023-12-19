import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { JobRunEvent } from '@neosync/sdk';
import { ReactElement, useMemo } from 'react';
import { getColumns } from './JobRunActivityTable/columns';
import { DataTable } from './JobRunActivityTable/data-table';

interface JobRunActivityTableProps {
  jobRunEvents?: JobRunEvent[];
}

export default function JobRunActivityTable(
  props: JobRunActivityTableProps
): ReactElement {
  const { jobRunEvents } = props;

  const columns = useMemo(() => getColumns({}), []);

  if (!jobRunEvents) {
    return <SkeletonTable />;
  }
  const isError = jobRunEvents.some((e) => e.tasks.some((t) => t.error));

  return (
    <div>
      <DataTable columns={columns} data={jobRunEvents} isError={isError} />
    </div>
  );
}
