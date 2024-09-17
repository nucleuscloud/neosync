import { cn } from '@/libs/utils';
import { format } from 'date-fns';
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

  // TODO: feed the subtasks to the the pop up and display them there

  return (
    <div
      className="w-full relative border border-gray-400 dark:border-gray-700 rounded overflow-hidden"
      style={{ height: `${tasks.length * 40 + 140}px` }} // 140 for the tooltip which we need to update and fix
    >
      {/* time axis */}
      <div className="sticky top-0 h-16 bg-gray-200 dark:bg-gray-800 z-10 px-6">
        <div className="relative w-full h-full">
          {timeLabels.map((label, index) => (
            <div
              key={index}
              className="absolute top-0 text-xs text-gray-600 dark:text-gray-300"
              style={{ left: `${getPositionPercentage(label)}%` }}
            >
              <div className="whitespace-nowrap py-1">{formatDate(label)}</div>
              <div className="whitespace-nowrap">{formatTime(label)}</div>
              <div className="h-4 w-[4px] rounded-full bg-gray-700 mx-auto" />
            </div>
          ))}
        </div>
      </div>
      {/* steps */}
      <div className="relative pt-12 ">
        {/* the grid lines */}
        {tasks.map((_, index) => (
          <div
            key={`grid-line-${index}`}
            className="absolute left-0 right-0 border-t border-gray-300 dark:border-gray-700"
            style={{ top: `${index * 40}px` }}
          />
        ))}
        {tasks.map((task, index) => {
          const left = getPositionPercentage(task.start);
          const width = getPositionPercentage(task.end) - left;

          return (
            <div
              key={task.id}
              className={cn(
                isError ? 'bg-red-400' : 'bg-blue-500',
                'absolute h-8 rounded hover:bg-blue-600 cursor-pointer mx-6 flex items-center'
              )}
              style={{
                left: `${left}%`,
                width: `${width}%`,
                top: `${index * 40 + 4}px`,
              }}
              onClick={() => onTaskClick?.(task)}
              onMouseEnter={() => setHoveredTask(task.id)}
              onMouseLeave={() => setHoveredTask(null)}
            >
              <div className="px-2 text-gray-900 dark:text-gray-200 text-sm w-full text-center justify-start flex">
                {task.name}
              </div>
              {hoveredTask === task.id && (
                <ActivityStepHover
                  task={task}
                  formatFullDate={formatFullDate}
                />
              )}
            </div>
          );
        })}
      </div>
    </div>
  );

  interface ActivityStepHoverProps {
    task: Task;
    formatFullDate: (date: Date) => string;
  }

  function ActivityStepHover(props: ActivityStepHoverProps): ReactElement {
    const { task, formatFullDate } = props;

    return (
      <div className="absolute top-full left-0 mt-2 p-2 bg-white dark:bg-gray-700 dark:border dark:border-gray-700 shadow-lg rounded z-20 whitespace-nowrap text-sm space-y-2">
        <div>
          <strong>Start:</strong>{' '}
          <Badge variant="default">{formatFullDate(task.start)}</Badge>
        </div>
        <div>
          <strong>Finish:</strong>{' '}
          <Badge variant="default">{formatFullDate(task.end)}</Badge>
        </div>
        {task.dependencies && (
          <p>
            <strong>Dependencies:</strong> {task.dependencies.join(', ')}
          </p>
        )}
      </div>
    );
  }
}
