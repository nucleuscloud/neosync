import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/libs/utils';
import { format, formatDuration, intervalToDuration } from 'date-fns';
import { ReactElement, useMemo, useState } from 'react';
import { Badge } from '../ui/badge';

interface Props {
  isError: boolean;
  tasks: Task[];
  onTaskClick?: (task: Task) => void;
}
export interface Task {
  id: string;
  name: string;
  start: Date;
  end: Date;
  dependencies?: string[];
}

export default function RunTimeline(props: Props): ReactElement {
  const { isError, tasks, onTaskClick } = props;

  const [hoveredTask, setHoveredTask] = useState<string | null>(null);

  const formatFullDate = (date: Date) => {
    return format(date, 'MM/dd/yyyy HH:mm:ss:SSS');
  };

  const formatDate = (date: Date) => format(date, 'MM/dd/yyyy');
  const formatTime = (date: Date) => format(date, 'HH:mm:ss:SSS');

  const formatTaskDuration = (start: Date, end: Date) => {
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
  const { timelineStart, totalDuration, timeLabels } = useMemo(() => {
    // find the first start time
    const start = new Date(Math.min(...tasks.map((t) => t.start.getTime())));
    // find the last end time
    const end = new Date(Math.max(...tasks.map((t) => t.end.getTime())));
    // find duration
    const duration = end.getTime() - start.getTime();

    // add some padding to the start and end to get an adjusted duration to get everything into view
    const adjustedStart = new Date(start.getTime() - duration * 0.05);
    const adjustedEnd = new Date(end.getTime() + duration * 0.05);
    const adjustedDuration = adjustedEnd.getTime() - adjustedStart.getTime();

    const labelCount = 6;
    const labels: Date[] = Array.from(
      { length: labelCount },
      (_, i) =>
        new Date(adjustedStart.getTime() + (adjustedDuration * i) / labelCount)
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
  return (
    <div
      className="w-full relative border border-gray-400 dark:border-gray-700 rounded overflow-hidden max-h-[800px]"
      // style={{ height: `${tasks.length * 40 + 200}px` }}
    >
      <div className="flex flex-row h-full">
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
                  <div>{task.name}</div>
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
            const left = getPositionPercentage(task.start);
            const width = getPositionPercentage(task.end) - left;

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
                        onClick={() => onTaskClick?.(task)}
                      >
                        <div className="px-2 text-gray-900 dark:text-gray-200 text-sm w-full flex flex-row gap-4 items-center">
                          <p>{task.name}</p>
                          <span className="text-xs bg-black text-white px-1 py-0.5 rounded text-nowrap">
                            {formatTaskDuration(task.start, task.end)}
                          </span>
                        </div>
                      </div>
                    </TooltipTrigger>
                    <TooltipContent align="start">
                      <div>
                        <strong>Start:</strong>{' '}
                        <Badge variant="default">
                          {formatFullDate(task.start)}
                        </Badge>
                      </div>
                      <div>
                        <strong>Finish:</strong>{' '}
                        <Badge variant="default">
                          {formatFullDate(task.end)}
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
      <div className="sticky top-0 h-14 border-b border-gray-300 dark:border-gray-700 bg-gray-200 dark:bg-gray-800 z-10 px-6">
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
