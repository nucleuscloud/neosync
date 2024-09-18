import RunTimeline, { Task } from '@/components/RunTImeline/RunTimeline';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { JobRunEvent } from '@neosync/sdk';
import { ReactElement, useMemo } from 'react';
import { getColumns } from './JobRunActivityTable/columns';
import { DataTable } from './JobRunActivityTable/data-table';

interface JobRunActivityTableProps {
  jobRunEvents?: JobRunEvent[];
  onViewSelectClicked(schema: string, table: string): void;
}

export default function JobRunActivityTable(
  props: JobRunActivityTableProps
): ReactElement {
  const { jobRunEvents, onViewSelectClicked } = props;

  const columns = useMemo(() => getColumns({}), []);

  if (!jobRunEvents) {
    return <SkeletonTable />;
  }
  const isError = jobRunEvents.some((e) => e.tasks.some((t) => t.error));

  return (
    <div className="flex flex-col gap-4">
      <RunTimeline
        isError={isError}
        tasks={convertJobRunEventToTask(jobRunEvents)}
      />
      <DataTable
        columns={columns}
        data={jobRunEvents}
        isError={isError}
        onViewSelectClicked={onViewSelectClicked}
      />
    </div>
  );
}

/*

this is what one object looks like
{
    "id": "5",
    "type": "RetrieveActivityOptions",
    "startTime": "2024-09-15T06:18:30.274318131Z",
    "closeTime": "2024-09-15T06:18:30.279649339Z",
    "tasks": [
        {
            "id": "5",
            "type": "ActivityTaskScheduled",
            "eventTime": "2024-09-15T06:18:30.271563631Z"
        },
        {
            "id": "6",
            "type": "ActivityTaskStarted",
            "eventTime": "2024-09-15T06:18:30.274318131Z"
        },
        {
            "id": "7",
            "type": "ActivityTaskCompleted",
            "eventTime": "2024-09-15T06:18:30.279649339Z"
        }
    ]
}
*/

function convertJobRunEventToTask(jre: JobRunEvent[]): Task[] {
  console.log('Input job run events:', jre);

  /* 
  some tasks, like the checkAccoutStatus will get scheudled but never run if the job finishes before it's set to run, and as a result, the JobRunEvent doesn't return a closeTime, so we need to filter these out so they don't continue to run in the table
  */

  const tasks = jre
    .filter(
      (event) => event.closeTime !== undefined && event.closeTime !== null
    )
    .flatMap((event: JobRunEvent, eventIndex: number): Task[] => {
      const mainTask: Task = {
        id: `event-${eventIndex}`,
        name: event.type,
        start: convertTimestampToDate(event.startTime),
        end: convertTimestampToDate(event.closeTime),
      };

      return [mainTask];
    });

  console.log('Converted tasks:', tasks);
  return tasks;
}

function convertTimestampToDate(
  timestamp: { seconds: bigint; nanos: number } | undefined
): Date {
  if (!timestamp) return new Date();

  const millisecondsFromSeconds = Number(timestamp.seconds) * 1000;
  const millisecondsFromNanos = timestamp.nanos / 1_000_000;

  const totalMilliseconds = millisecondsFromSeconds + millisecondsFromNanos;

  return new Date(totalMilliseconds);
}
