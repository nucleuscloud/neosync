import { ZoomInIcon, ZoomOutIcon } from '@radix-ui/react-icons';
import { format } from 'date-fns';
import { ReactElement, useEffect, useMemo, useRef, useState } from 'react';
import { Button } from '../ui/button';

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
  const { tasks, onTaskClick } = props;

  const [hoveredTask, setHoveredTask] = useState<string | null>(null);
  const [zoomLevel, setZoomLevel] = useState(1);
  const [containerWidth, setContainerWidth] = useState(0);
  const timelineRef = useRef<HTMLDivElement>(null);

  // console.log('tasks', tasks, isError);

  useEffect(() => {
    const updateWidth = () => {
      if (timelineRef.current) {
        setContainerWidth(timelineRef.current.offsetWidth);
      }
    };

    updateWidth();
    window.addEventListener('resize', updateWidth);

    return () => window.removeEventListener('resize', updateWidth);
  }, []);
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
    const padding = duration * 0.05 * (1 / zoomLevel);
    const adjustedStart = new Date(start.getTime() - padding);
    const adjustedEnd = new Date(end.getTime() + padding);
    const adjustedDuration = adjustedEnd.getTime() - adjustedStart.getTime();

    const labelCount = 4;
    const labels: Date[] = Array.from(
      { length: labelCount + 1 },
      (_, i) =>
        new Date(adjustedStart.getTime() + (adjustedDuration * i) / labelCount)
    );

    return {
      timelineStart: adjustedStart,
      timelineEnd: adjustedEnd,
      totalDuration: adjustedDuration,
      timeLabels: labels,
    };
  }, [tasks, zoomLevel]);

  // calculates where in the timeline axis something should be relative to the total duration
  const getPositionPixels = (time: Date) => {
    const percentage =
      (time.getTime() - timelineStart.getTime()) / totalDuration;
    return percentage * containerWidth * zoomLevel;
  };
  const handleZoomIn = () => setZoomLevel((prev) => Math.min(prev * 1.5, 10));
  const handleZoomOut = () => setZoomLevel((prev) => Math.max(prev / 1.5, 0.1));

  return (
    <div className="flex flex-col gap-4 pt-4">
      <div className="flex gap-2 justify-end">
        <Button onClick={handleZoomIn} variant="outline">
          <ZoomInIcon />
        </Button>
        <Button onClick={handleZoomOut} variant="outline">
          <ZoomOutIcon />
        </Button>
      </div>
      <div
        className="w-full h-[500px] relative bg-white border border-gray-800 rounded"
        id="timeline"
        ref={timelineRef}
      >
        <div className="sticky top-0  right-0 h-16 bg-gray-200 z-10 ">
          <div
            className="relative w-full h-full"
            style={{ width: `${100 * zoomLevel}%` }}
          >
            {timeLabels.map((label, index) => {
              const position = getPositionPixels(label);
              return (
                <div
                  key={index}
                  className="absolute top-0 text-xs text-gray-600 flex flex-col items-center"
                  style={{ left: `${position}px` }}
                >
                  <div className="transform -translate-x-1/2 whitespace-nowrap py-1">
                    {formatDate(label)}
                  </div>
                  <div className="transform -translate-x-1/2 whitespace-nowrap">
                    {formatTime(label)}
                  </div>
                  <div className="h-2 w-px bg-gray-300 mt-1" />
                </div>
              );
            })}
          </div>
        </div>
        <div className="relative pt-16 overflow-x-auto overflow-y-auto h-[calc(100%-4rem)]">
          <div style={{ width: `${100 * zoomLevel}%`, position: 'relative' }}>
            {tasks.map((task, index) => {
              const left = getPositionPixels(task.start);
              const width = getPositionPixels(task.end) - left;

              return (
                <div
                  key={task.id}
                  className="absolute h-8 bg-blue-500 rounded hover:bg-blue-600 cursor-pointer"
                  style={{
                    left: `${left}px`,
                    width: `${width}px`,
                    top: `${index * 40}px`,
                  }}
                  onClick={() => onTaskClick?.(task)}
                  onMouseEnter={() => setHoveredTask(task.id)}
                  onMouseLeave={() => setHoveredTask(null)}
                >
                  <div className="px-2 text-black">{task.name}</div>
                  {hoveredTask === task.id && (
                    <div className="absolute top-full left-0 mt-2 p-2 bg-white shadow-lg rounded z-20 whitespace-nowrap">
                      <p>
                        <strong>Start:</strong> {formatDate(task.start)}{' '}
                        {formatTime(task.start)}
                      </p>
                      <p>
                        <strong>End:</strong> {formatDate(task.end)}{' '}
                        {formatTime(task.end)}
                      </p>
                      {task.dependencies && (
                        <p>
                          <strong>Dependencies:</strong>{' '}
                          {task.dependencies.join(', ')}
                        </p>
                      )}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}
