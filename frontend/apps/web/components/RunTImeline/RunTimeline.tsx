import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/libs/utils';
import { Timestamp } from '@bufbuild/protobuf';
import { JobRunEvent, JobRunStatus } from '@neosync/sdk';

import { JobRunStatus as JobRunStatusEnum } from '@neosync/sdk';
import {
  CheckCircledIcon,
  CrossCircledIcon,
  MinusCircledIcon,
  MixerHorizontalIcon,
} from '@radix-ui/react-icons';
import {
  addMilliseconds,
  format,
  formatDuration,
  intervalToDuration,
} from 'date-fns';
import { ReactElement, useMemo, useState } from 'react';
import Spinner from '../Spinner';
import TruncatedText from '../TruncatedText';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';

interface Props {
  tasks: JobRunEvent[];
  jobStatus?: JobRunStatusEnum;
}

export default function RunTimeline(props: Props): ReactElement {
  const { tasks, jobStatus } = props;

  // this should probably be better typed and using the ActivityStatus types
  // but those aren't currenlty sent to the FE as a status but are represented in the
  // type field
  const [selectedStatuses, setSelectedStatuses] = useState<string[]>([
    'running',
    'completed',
    'failed',
    'canceled',
  ]);

  const formatFullDate = (date: Timestamp | Date | undefined) => {
    if (!date) return 'N/A';

    if (date instanceof Timestamp) {
      return format(convertTimestampToDate(date), 'MM/dd/yyyy HH:mm:ss:SSS');
    }

    if (date instanceof Date) {
      return format(date, 'MM/dd/yyyy HH:mm:ss:SSS');
    }
  };

  const formatDate = (date: Date) => format(date, 'MM/dd/yyyy');
  const formatTime = (date: Date) => format(date, 'HH:mm:ss:SSS');

  const formatTaskDuration = (
    s: Timestamp | undefined,
    end: Date | undefined
  ) => {
    if (!s || !end) return 'N/A';
    const start = convertTimestampToDate(s);
    const duration = intervalToDuration({ start, end });
    const milliseconds = end.getTime() - start.getTime();
    const millis = milliseconds % 1000;

    // format the duration string
    const formattedDuration = formatDuration(duration, {
      format: ['hours', 'minutes', 'seconds'],
      delimiter: ', ',
    });

    if (!formattedDuration) {
      return `${millis} ms`;
    }

    return `${formattedDuration}${millis > 0 ? `, ${millis} ms` : ''}`;
  };

  const { timelineStart, totalDuration, timeLabels } = useMemo(() => {
    // find earliest start date out of all of the activities
    let startTime = Infinity;
    let endTime = -Infinity;

    tasks.map((t) => {
      startTime = Math.min(
        startTime,
        convertTimestampToDate(t.startTime).getTime()
      );

      const errorDate = getCloseOrErrorOrCancelDate(t);
      endTime = Math.max(
        endTime,
        errorDate.getTime(),
        convertTimestampToDate(t.closeTime || t.startTime).getTime()
      );
    });

    const start = new Date(startTime);
    const end = new Date(endTime);

    let duration = end.getTime() - start.getTime();

    // add padding, but limit it to a maximum of 100ms on each side so we can see the entire timeline in view in the graph
    const padding = Math.min(duration * 0.1, 300);
    const adjustedStart = new Date(start.getTime() - padding);
    const adjustedEnd = new Date(end.getTime() + padding);
    const adjustedDuration = adjustedEnd.getTime() - adjustedStart.getTime();

    const labelCount = 5;

    const labels: Date[] = Array.from({ length: labelCount }, (_, i) =>
      addMilliseconds(adjustedStart, (adjustedDuration * i) / (labelCount - 1))
    );

    return {
      timelineStart: adjustedStart,
      timelineEnd: adjustedEnd,
      totalDuration: adjustedDuration,
      timeLabels: labels,
    };
  }, [tasks]);

  // calculates where in the timeline axis something should be relative to the total duration
  // also dictates how far right the timeline goes, reduce if you want the timeline shorter or length otherwise
  const getPositionPercentage = (time: Date) => {
    return ((time.getTime() - timelineStart.getTime()) / totalDuration) * 92;
  };

  // handles getting the activity statuses by remaping the ActivityStatuses since (i think) we don't want to show all of them per activity
  // prob want to update with actual types instead of using strings here
  // the event types in the types field are just the stringified Temporal types
  const getTaskStatus = (task: JobRunEvent): string => {
    const hasCompleted = task.tasks.some(
      (item) => item.type === 'ActivityTaskCompleted'
    );
    const hasFailed = task.tasks.some(
      (item) => item.type === 'ActivityTaskFailed' || item.error
    );
    const isCanceled = task.tasks.some(
      (item) => item.type === 'ActivityTaskCancelRequested'
    );

    const isJobTerminated = jobStatus && jobStatus == JobRunStatus.TERMINATED;

    if (hasCompleted) return 'completed';
    if (hasFailed) return 'failed';
    if (isCanceled || (isJobTerminated && !hasCompleted)) return 'canceled';
    if (!hasCompleted && !hasFailed && !isCanceled) return 'running';
    return 'unknown';
  };

  // handles filtering the tasks when the tasks or filters change
  const filteredTasks = useMemo(() => {
    return tasks.filter((task) => {
      const status = getTaskStatus(task);
      return selectedStatuses.includes(status);
    });
  }, [tasks, selectedStatuses, jobStatus]);

  const handleStatusFilterChange = (status: string, checked: boolean) => {
    setSelectedStatuses((prev) =>
      checked ? [...prev, status] : prev.filter((s) => s !== status)
    );
  };

  console.log('tasks', tasks);
  return (
    <div className="flex flex-col gap-2">
      <div className="flex justify-between w-full">
        <div className="text-xl font-semibold">Activity Timeline</div>
        <StatusFilter
          selectedStatuses={selectedStatuses}
          onStatusChange={handleStatusFilterChange}
        />
      </div>
      <div className="w-full relative border border-gray-200 dark:border-gray-700 rounded overflow-y-auto max-h-[400px]">
        <div className="flex flex-row h-full w-full">
          {/* the left activity bar */}
          <div className="w-1/6">
            <div className="sticky top-0 h-14 bg-gray-200 dark:bg-gray-800 z-10 px-6 border-b border-gray-200 dark:border-gray-700" />
            <div className="border-r border-gray-200 dark:border-gray-700 flex flex-col text-sm ">
              {filteredTasks.map((task, index) => {
                const isLastItem = index === tasks.length - 1;
                return (
                  <div
                    key={task.id}
                    className={cn(
                      'px-2 h-10 items-center flex',
                      !isLastItem &&
                        'border-b border-gray-200 dark:border-gray-700'
                    )}
                  >
                    <ActivityLabel task={task} getTaskStatus={getTaskStatus} />
                  </div>
                );
              })}
            </div>
          </div>
          <div className="relative w-5/6">
            <TableHeader
              getPositionPercentage={getPositionPercentage}
              formatDate={formatDate}
              timeLabels={timeLabels}
            />

            {filteredTasks.map((_, index) => (
              <div
                key={`grid-line-${index}`}
                className="absolute left-0 right-0 border-t border-gray-200 dark:border-gray-700"
                style={{ top: `${index * 40 + 55}px` }}
                id="grid-lines"
              />
            ))}
            {filteredTasks.map((task, index) => {
              const failedTask = task.tasks.find((item) => item.error);

              const left = getPositionPercentage(
                convertTimestampToDate(task.startTime)
              );
              const endTime = getCloseOrErrorOrCancelDate(task);
              const width = getPositionPercentage(endTime) - left;
              const status = getTaskStatus(task);

              const cancelTime = task.tasks.find(
                (t) => t.type === 'ActivityTaskCancelRequested'
              )?.eventTime;
              console.log('cancel Time', cancelTime, task);

              return (
                <div className="flex flex-row" key={task.id}>
                  <TooltipProvider delayDuration={100}>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <div
                          className={cn(
                            status === 'failed'
                              ? 'bg-red-400 dark:bg-red-700'
                              : status === 'canceled'
                                ? 'bg-yellow-400 dark:bg-yellow-700'
                                : 'bg-blue-500',
                            'absolute h-8 rounded hover:bg-opacity-80 cursor-pointer mx-6 flex items-center'
                          )}
                          style={{
                            left: `${left}%`,
                            width: `${width}%`,
                            top: `${index * 40 + 60}px`,
                          }}
                        >
                          <div className="px-2 text-gray-900 dark:text-gray-200 text-sm w-full flex flex-row gap-4 items-center">
                            <span className="text-xs bg-black dark:bg-gray-700 text-white px-1 py-0.5 rounded text-nowrap">
                              {formatTaskDuration(task.startTime, endTime)}
                            </span>
                            <SyncLabel task={task} />
                          </div>
                        </div>
                      </TooltipTrigger>
                      <TooltipContent
                        align="start"
                        className="dark:bg-gray-800 shadow-lg border dark:border-gray-700 flex flex-col gap-1"
                      >
                        <div className="flex flex-row gap-2 items-center justify-between w-full">
                          <strong>Start:</strong>{' '}
                          <Badge variant="default" className="w-[180px]">
                            {formatFullDate(task.startTime)}
                          </Badge>
                        </div>
                        <div className="flex flex-row gap-2 items-center justify-between w-full">
                          <strong>Finish:</strong>{' '}
                          <Badge variant="default" className="w-[180px]">
                            {status == 'failed' ||
                            status == 'terminated' ||
                            status == 'canceled'
                              ? 'N/A'
                              : formatFullDate(endTime)}
                          </Badge>
                        </div>
                        {failedTask && (
                          <div className="flex flex-row gap-2 justify-between w-full">
                            <strong>Error:</strong>{' '}
                            <Badge variant="destructive" className="w-[180px]">
                              {failedTask.error?.message || 'Unknown error'}
                            </Badge>
                          </div>
                        )}
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );

  interface TableHeaderProps {
    formatDate: (date: Date) => string;
    getPositionPercentage: (time: Date) => number;
    timeLabels: Date[];
  }

  function TableHeader(props: TableHeaderProps): ReactElement {
    const { formatDate, getPositionPercentage, timeLabels } = props;

    return (
      <div className="w-full sticky top-0 h-14 border-b border-gray-200 dark:border-gray-700 bg-gray-200 dark:bg-gray-800 z-10 ">
        <div className="relative w-full h-full">
          {timeLabels.map((label, index) => (
            <div
              key={index}
              className="absolute top-0 text-xs text-gray-700 dark:text-gray-300"
              style={{ left: `${getPositionPercentage(label)}%` }}
            >
              <div className="whitespace-nowrap py-1">{formatDate(label)}</div>
              <div className="whitespace-nowrap">{formatTime(label)}</div>
              <div className="h-4 w-[1px] rounded-full bg-gray-500 mx-auto" />
            </div>
          ))}
        </div>
      </div>
    );
  }
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

function getCloseOrErrorOrCancelDate(task: JobRunEvent): Date {
  const errorTask = task.tasks.find((item) => item.error);
  const errorTime = errorTask ? errorTask.eventTime : undefined;
  const cancelTime = task.tasks.find(
    (t) => t.type === 'ActivityTaskCancelRequested'
  )?.eventTime;
  return errorTime
    ? convertTimestampToDate(errorTime)
    : cancelTime
      ? convertTimestampToDate(cancelTime)
      : convertTimestampToDate(task.closeTime);
}

interface SyncLabelProps {
  task: JobRunEvent;
}

function SyncLabel(props: SyncLabelProps) {
  const { task } = props;

  const schemaTable = `${task.metadata?.metadata.value?.schema}.${task.metadata?.metadata.value?.table} `;

  return (
    <div className="text-white">
      {task.metadata?.metadata.case == 'syncMetadata' && schemaTable}
    </div>
  );
}

interface ActivityLabelProps {
  task: JobRunEvent;
  getTaskStatus: (task: JobRunEvent) => string;
}

function ActivityLabel({ task, getTaskStatus }: ActivityLabelProps) {
  const status = getTaskStatus(task);

  const handleStatus = () => {
    switch (status) {
      case 'completed':
        return <CheckCircledIcon className="text-green-500" />;
      case 'failed':
        return <CrossCircledIcon className="text-red-500" />;
      case 'canceled':
        return <MinusCircledIcon className="text-yellow-500" />;
      case 'running':
        return <Spinner />;
      default:
        return null;
    }
  };

  return (
    <div className="flex flex-row items-center gap-2">
      {task.id.toString()}.
      <TruncatedText text={task.type} />
      {handleStatus()}
    </div>
  );
}

interface StatusFilterProps {
  selectedStatuses: string[];
  onStatusChange: (status: string, checked: boolean) => void;
}

// would be nice to replace with a multi-select so you don't have to open/close it everytime you want to make a change
function StatusFilter({ selectedStatuses, onStatusChange }: StatusFilterProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">
          {' '}
          <MixerHorizontalIcon className="mr-2 h-4 w-4" />
          Status
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-56">
        <DropdownMenuCheckboxItem
          checked={selectedStatuses.includes('running')}
          onCheckedChange={(checked) => onStatusChange('running', checked)}
        >
          Running
        </DropdownMenuCheckboxItem>
        <DropdownMenuCheckboxItem
          checked={selectedStatuses.includes('completed')}
          onCheckedChange={(checked) => onStatusChange('completed', checked)}
        >
          Completed
        </DropdownMenuCheckboxItem>
        <DropdownMenuCheckboxItem
          checked={selectedStatuses.includes('failed')}
          onCheckedChange={(checked) => onStatusChange('failed', checked)}
        >
          Failed
        </DropdownMenuCheckboxItem>
        <DropdownMenuCheckboxItem
          checked={selectedStatuses.includes('canceled')}
          onCheckedChange={(checked) => onStatusChange('canceled', checked)}
        >
          Canceled
        </DropdownMenuCheckboxItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
