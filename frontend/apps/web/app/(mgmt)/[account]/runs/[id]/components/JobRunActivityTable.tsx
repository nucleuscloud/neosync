import RunTimeline from '@/components/RunTImeline/RunTimeline';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { JobRunEvent, JobRunStatus as JobRunStatusEnum } from '@neosync/sdk';
import { ReactElement, useMemo } from 'react';
import { getColumns } from './JobRunActivityTable/columns';
import { DataTable } from './JobRunActivityTable/data-table';

interface JobRunActivityTableProps {
  jobRunEvents?: JobRunEvent[];
  onViewSelectClicked(schema: string, table: string): void;
  jobStatus?: JobRunStatusEnum;
}

export default function JobRunActivityTable(
  props: JobRunActivityTableProps
): ReactElement<any> {
  const { jobRunEvents, onViewSelectClicked, jobStatus } = props;

  const columns = useMemo(() => getColumns({ onViewSelectClicked }), []);

  if (!jobRunEvents) {
    return <SkeletonTable />;
  }
  const isError = jobRunEvents.some((e) => e.tasks.some((t) => t.error));

  return (
    <div className="flex flex-col gap-4">
      <RunTimeline tasks={jobRunEvents} jobStatus={jobStatus} />
      <div className="text-xl font-semibold">Activity Table</div>
      <DataTable
        columns={columns}
        data={jobRunEvents}
        isError={isError}
        onViewSelectClicked={onViewSelectClicked}
      />
    </div>
  );
}
