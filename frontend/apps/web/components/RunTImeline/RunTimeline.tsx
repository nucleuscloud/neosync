import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/libs/utils';
import { Timestamp } from '@bufbuild/protobuf';
import { JobRunEvent } from '@neosync/sdk';
import {
  addMilliseconds,
  format,
  formatDuration,
  intervalToDuration,
} from 'date-fns';
import { ReactElement, useMemo, useState } from 'react';
import { Badge } from '../ui/badge';

interface Props {
  isError: boolean;
  tasks: JobRunEvent[];
  onTaskClick?: (task: JobRunEvent) => void;
}

export default function RunTimeline(props: Props): ReactElement {
  const { isError, tasks, onTaskClick } = props;

  const [hoveredTask, setHoveredTask] = useState<string | null>(null);

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

    // Format the duration string without empty components
    const formattedDuration = formatDuration(duration, {
      format: ['hours', 'minutes', 'seconds'],
      delimiter: ', ',
    });

    if (!formattedDuration) {
      return `${millis} ms`;
    }

    return `${formattedDuration}${millis > 0 ? `, ${millis} ms` : ''}`;
  };
  const { timelineStart, timelineEnd, totalDuration, timeLabels } =
    useMemo(() => {
      const start = new Date(
        Math.min(
          ...tasks.map((t) => convertTimestampToDate(t.startTime).getTime())
        )
      );
      const end = new Date(
        Math.max(
          ...tasks.map((t) => {
            const errorDate = getCloseOrErrorDate(t);
            return Math.max(
              errorDate.getTime(),
              convertTimestampToDate(t.closeTime || t.startTime).getTime()
            );
          })
        )
      );

      let duration = end.getTime() - start.getTime();

      // Add padding, but limit it to a maximum of 100ms on each side
      const padding = Math.min(duration * 0.1, 100);
      const adjustedStart = new Date(start.getTime() - padding);
      const adjustedEnd = new Date(end.getTime() + padding);
      const adjustedDuration = adjustedEnd.getTime() - adjustedStart.getTime();

      const labelCount = 5;
      const labels: Date[] = Array.from({ length: labelCount }, (_, i) =>
        addMilliseconds(
          adjustedStart,
          (adjustedDuration * i) / (labelCount - 1)
        )
      );

      return {
        timelineStart: adjustedStart,
        timelineEnd: adjustedEnd,
        totalDuration: adjustedDuration,
        timeLabels: labels,
      };
    }, [tasks]);

  // calculates where in the timeline axis something should be relative to the total duration
  const getPositionPercentage = (time: Date) => {
    return ((time.getTime() - timelineStart.getTime()) / totalDuration) * 100;
  };

  console.log('tasks', tasks);
  console.log('isError', isError);

  // TODO: only show
  return (
    <div
      className="w-full relative border border-gray-400 dark:border-gray-700 rounded overflow-hidden  max-h-[800px]"
      // style={{ height: `${tasks.length * 40 + 200}px` }}
    >
      <div className="flex flex-row h-full w-full">
        <div className="w-1/6">
          <div className="sticky top-0 h-14 bg-gray-200 dark:bg-gray-800 z-10 px-6 border-b border-gray-300 dark:border-gray-700" />
          <div className="border-r border-gray-300 dark:border-gray-700 flex flex-col h-full text-sm ">
            {tasks.map((task, index) => {
              const isLastItem = index === tasks.length - 1;
              return (
                <div
                  key={task.id}
                  className={cn(
                    'px-2 h-10 items-center flex',
                    !isLastItem &&
                      'border-b border-gray-300 dark:border-gray-700'
                  )}
                >
                  <div>{task.type}</div>
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

          {tasks.map((_, index) => (
            <div
              key={`grid-line-${index}`}
              className="absolute left-0 right-0 border-t border-gray-300 dark:border-gray-700"
              style={{ top: `${index * 40 + 55}px` }}
              id="grid-lines"
            />
          ))}
          {tasks.map((task, index) => {
            const left = getPositionPercentage(
              convertTimestampToDate(task.startTime)
            );

            const endTime = getCloseOrErrorDate(task);
            const width = getPositionPercentage(endTime) - left;
            const errorTask = task.tasks.find((item) => item.error);

            // TODO: only highlight in red the task that failed
            return (
              <div className="flex flex-row">
                <TooltipProvider delayDuration={100}>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <div
                        key={task.id}
                        className={cn(
                          isError ? 'bg-red-400' : 'bg-blue-500',
                          'absolute h-8 rounded hover:bg-blue-600 cursor-pointer mx-6 flex items-center'
                        )}
                        style={{
                          left: `${left}%`,
                          width: `${width}%`,
                          top: `${index * 40 + 60}px`,
                        }}
                        // onClick={() => onTaskClick?.(task)}
                      >
                        <div className="px-2 text-gray-900 dark:text-gray-200 text-sm w-full flex flex-row gap-4 items-center">
                          <p>{task.type}</p>
                          <span className="text-xs bg-black text-white px-1 py-0.5 rounded text-nowrap">
                            {formatTaskDuration(task.startTime, endTime)}
                          </span>
                        </div>
                      </div>
                    </TooltipTrigger>
                    <TooltipContent align="start">
                      <div>
                        <strong>Start:</strong>{' '}
                        <Badge variant="default">
                          {formatFullDate(task.startTime)}
                        </Badge>
                      </div>
                      <div>
                        <strong>Finish:</strong>{' '}
                        <Badge variant="default">
                          {formatFullDate(endTime)}
                        </Badge>
                      </div>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
            );
          })}
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
      <div className="w-full sticky top-0 h-14 border-b border-gray-300 dark:border-gray-700 bg-gray-200 dark:bg-gray-800 z-10 ">
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

function getCloseOrErrorDate(task: JobRunEvent): Date {
  const errorTask = task.tasks.find((item) => item.error);
  const errorTime = errorTask ? errorTask.eventTime : undefined;
  return errorTime
    ? convertTimestampToDate(errorTime)
    : convertTimestampToDate(task.closeTime);
}
