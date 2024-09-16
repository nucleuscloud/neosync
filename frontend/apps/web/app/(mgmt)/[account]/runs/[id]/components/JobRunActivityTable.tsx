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
    <div>
      <DataTable
        columns={columns}
        data={jobRunEvents}
        isError={isError}
        onViewSelectClicked={onViewSelectClicked}
      />
      <RunTimeline
        isError={isError}
        tasks={convertJobRunEventToTask(jobRunEvents)}
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
  // jre might have sub tasks so we want to flatten the entire thing so we can use it in the timeline view with dependencies

  return jre.flatMap((jre: JobRunEvent, eventIndex: number): Task[] => {
    const mainTask: Task = {
      id: `event-${eventIndex}`,
      name: jre.type,
      start: convertTimestampToDate(jre.startTime),
      end: convertTimestampToDate(jre.closeTime),
    };

    // const subTasks: Task[] = jre.tasks.map(
    //   (task: JobRunEventTask, taskIndex: number) => ({
    //     id: `event-${eventIndex}-task-${taskIndex}`,
    //     name: task.type || `Subtask ${taskIndex + 1}`,
    //     start: task.eventTime
    //       ? new Date(Number(task.eventTime.seconds) * 10)
    //       : mainTask.start,
    //     end: task.eventTime
    //       ? new Date(Number(task.eventTime.seconds) * 30)
    //       : mainTask.end,
    //     dependencies: [mainTask.id],
    //   })
    // );

    // // If there are subtasks, make the main task depend on all subtasks
    // if (subTasks.length > 0) {
    //   mainTask.dependencies = subTasks.map((task) => task.id);
    // }

    return [mainTask];
  });
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
