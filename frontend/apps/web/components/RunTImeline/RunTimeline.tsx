import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { cn } from '@/libs/utils';
import {
  Timestamp,
  timestampDate,
  TimestampSchema,
} from '@bufbuild/protobuf/wkt';
import { JobRunEvent, JobRunStatus } from '@neosync/sdk';

import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { isMessage } from '@bufbuild/protobuf';
import { JobRunStatus as JobRunStatusEnum } from '@neosync/sdk';
import {
  CheckCircledIcon,
  ChevronDownIcon,
  ChevronUpIcon,
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
import React, { ReactElement, useMemo, useState } from 'react';
import Spinner from '../Spinner';
import TruncatedText from '../TruncatedText';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';

interface Props {
  tasks: JobRunEvent[];
  jobStatus?: JobRunStatusEnum;
}

const expandedRowHeight = 165;
const defaultRowHeight = 40;

type RunStatus = 'running' | 'completed' | 'failed' | 'canceled' | 'terminated';

export default function RunTimeline(props: Props): ReactElement<any> {
  const { tasks, jobStatus } = props;
  const [expandedTaskId, setExpandedTaskId] = useState<string | null>(null);
  const [selectedStatuses, setSelectedStatuses] = useState<RunStatus[]>([
    'running',
    'completed',
    'failed',
    'canceled',
    'terminated',
  ]);

  const { timelineStart, totalDuration, timeLabels } = useMemo(() => {
    // find earliest start date out of all of the activities
    let startTime = Infinity;
    let endTime = -Infinity;

    tasks.forEach((t) => {
      const scheduled = t.tasks.find(
        (st) =>
          st.type === 'ActivityTaskScheduled' ||
          st.type === 'StartChildWorkflowExecutionInitiated'
      )?.eventTime;
      startTime = Math.min(
        startTime,
        convertTimestampToDate(scheduled).getTime()
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

  // handles filtering the tasks when the tasks or filters change
  const filteredTasks = useMemo(() => {
    return tasks.filter((task) => {
      const status = getTaskStatus(task, jobStatus);
      return selectedStatuses.includes(status);
    });
  }, [tasks, selectedStatuses, jobStatus]);

  function handleStatusFilterChange(status: RunStatus, checked: boolean) {
    setSelectedStatuses((prev) =>
      checked ? [...prev, status] : prev.filter((s) => s !== status)
    );
  }

  function toggleExpandedRowBody(taskId: string) {
    setExpandedTaskId((prevId) => (prevId === taskId ? null : taskId));
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex justify-between w-full">
        <div className="text-xl font-semibold">Activity Timeline</div>
        <StatusFilter
          selectedStatuses={selectedStatuses}
          onStatusChange={handleStatusFilterChange}
        />
      </div>
      <div className="w-full relative border border-gray-200 dark:border-gray-700 rounded overflow-y-scroll max-h-[400px]">
        <div className="flex flex-row h-full w-full">
          <div className="w-1/6">
            <LeftActivityBar
              filteredTasks={filteredTasks}
              toggleExpandedRowBody={toggleExpandedRowBody}
              jobStatus={jobStatus}
              expandedTaskId={expandedTaskId ?? ''}
            />
          </div>
          <div className="relative w-5/6">
            <TableHeader
              getPositionPercentage={getPositionPercentage}
              formatDate={formatDate}
              timeLabels={timeLabels}
              timelineStart={timelineStart}
              totalDuration={totalDuration}
            />
            {filteredTasks.map((task, index) => {
              const isExpanded = expandedTaskId === String(task.id);
              const isLastItem = index === filteredTasks.length - 1;
              // calcs an offset for the other rows to slide down so everything stays aligned
              const expandedOffset = filteredTasks
                .slice(0, index)
                .reduce(
                  (acc, t) =>
                    acc +
                    (expandedTaskId === String(t.id)
                      ? expandedRowHeight - defaultRowHeight
                      : 0),
                  0
                );
              // offset for the top header
              const topOffset = index * defaultRowHeight + 55 + expandedOffset;
              return (
                <React.Fragment key={task.id}>
                  <div
                    className="absolute left-0 right-0 border-t border-gray-200 dark:border-gray-700"
                    style={{ top: `${topOffset}px` }}
                    id="grid-lines"
                  />
                  <TimelineBar
                    task={task}
                    index={index}
                    jobStatus={jobStatus}
                    timelineStart={timelineStart}
                    totalDuration={totalDuration}
                    topOffset={topOffset}
                    expandedTaskId={expandedTaskId}
                    toggleExpandedRowBody={toggleExpandedRowBody}
                    isExpanded={isExpanded}
                    isLastItem={isLastItem}
                  />
                </React.Fragment>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}

interface LeftActivityBarProps {
  filteredTasks: JobRunEvent[];
  toggleExpandedRowBody: (val: string) => void;
  jobStatus: JobRunStatus | undefined;
  expandedTaskId: string;
}

function LeftActivityBar(props: LeftActivityBarProps): ReactElement<any> {
  const { filteredTasks, toggleExpandedRowBody, jobStatus, expandedTaskId } =
    props;
  return (
    <div>
      <div className="sticky top-0 h-14 bg-gray-200 dark:bg-gray-800 z-10 px-6 border-b border-gray-200 dark:border-gray-700" />
      <div className="border-r border-gray-200 dark:border-gray-700 flex flex-col text-sm">
        {filteredTasks.map((task, index) => {
          const isLastItem = index === filteredTasks.length - 1;
          const isExpanded = expandedTaskId === String(task.id);
          return (
            <div
              className={cn(
                'px-2 h-10 items-center flex cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700',
                !isLastItem && 'border-b border-gray-200 dark:border-gray-700',
                isExpanded && 'h-[165px]'
              )}
              onClick={() => toggleExpandedRowBody(String(task.id))}
              key={task.id}
            >
              <ActivityLabel
                task={task}
                status={getTaskStatus(task, jobStatus)}
              />
            </div>
          );
        })}
      </div>
    </div>
  );
}

interface TimelineBarProps {
  index: number;
  timelineStart: Date;
  totalDuration: number;
  task: JobRunEvent;
  jobStatus: JobRunStatus | undefined;
  topOffset: number;
  expandedTaskId: string | null;
  toggleExpandedRowBody: (val: string) => void;
  isExpanded: boolean;
  isLastItem: boolean;
}

function TimelineBar(props: TimelineBarProps) {
  const {
    task,
    jobStatus,
    timelineStart,
    totalDuration,
    topOffset,
    toggleExpandedRowBody,
    isExpanded,
    isLastItem,
  } = props;

  const scheduled = task.tasks.find(
    (st) =>
      st.type == 'ActivityTaskScheduled' ||
      st.type == 'StartChildWorkflowExecutionInitiated'
  )?.eventTime;

  const failedTask = task.tasks.find((item) => item.error);
  const left = getPositionPercentage(
    convertTimestampToDate(scheduled),
    timelineStart,
    totalDuration
  );
  const endTime = getCloseOrErrorOrCancelDate(task);
  const width =
    getPositionPercentage(endTime, timelineStart, totalDuration) - left;
  const status = getTaskStatus(task, jobStatus);

  return (
    <TooltipProvider delayDuration={100}>
      <Tooltip>
        <TooltipTrigger asChild>
          <div>
            <div
              className={cn(
                status === 'failed'
                  ? 'bg-red-400 dark:bg-red-700'
                  : status === 'canceled' || status === 'terminated'
                    ? 'bg-yellow-400 dark:bg-yellow-700'
                    : 'bg-blue-500',
                'absolute h-8 rounded hover:bg-opacity-80 cursor-pointer mx-6 flex items-center '
              )}
              style={{
                left: `${left}%`,
                width: `${width}%`,
                top: `${topOffset + 5}px`,
              }}
            >
              <div
                className="px-2 text-gray-900 dark:text-gray-200 text-sm w-full flex flex-row gap-4 items-center "
                onClick={() => toggleExpandedRowBody(String(task.id))}
              >
                <span className="text-xs bg-black dark:bg-gray-700 text-white px-1 py-0.5 rounded text-nowrap">
                  {formatTaskDuration(scheduled, endTime)}
                </span>
                <SyncLabel task={task} />
              </div>
            </div>
            <ExpandedRow
              toggleExpandedRowBody={toggleExpandedRowBody}
              isLastItem={isLastItem}
              isExpanded={isExpanded}
              task={task}
            />
          </div>
        </TooltipTrigger>
        <TooltipContent
          align="center"
          className="dark:bg-gray-800 shadow-lg border dark:border-gray-700 flex flex-col gap-1"
        >
          {isSyncActivity(task) && (
            <div className="flex flex-row gap-2 items-center justify-between w-full">
              <strong>Table:</strong>{' '}
              <Badge variant="default" className="w-[180px]">
                {}
                <SyncLabel task={task} />
              </Badge>
            </div>
          )}
          <div className="flex flex-row gap-2 items-center justify-between w-full">
            <strong>Start:</strong>{' '}
            <Badge variant="default" className="w-[180px]">
              {formatFullDate(scheduled)}
            </Badge>
          </div>
          <div className="flex flex-row gap-2 items-center justify-between w-full">
            <strong>Finish:</strong>{' '}
            <Badge variant="default" className="w-[180px]">
              {status == 'failed' ||
              status == 'canceled' ||
              status == 'terminated'
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
  );
}

interface TableHeaderProps {
  formatDate: (date: Date) => string;
  getPositionPercentage: (
    time: Date,
    timelineStart: Date,
    totalDuration: number
  ) => number;
  timeLabels: Date[];
  timelineStart: Date;
  totalDuration: number;
}

function TableHeader(props: TableHeaderProps): ReactElement<any> {
  const {
    formatDate,
    getPositionPercentage,
    timeLabels,
    timelineStart,
    totalDuration,
  } = props;

  return (
    <div className="w-full sticky top-0 h-14 border-b border-gray-200 dark:border-gray-700 bg-gray-200 dark:bg-gray-800 z-10 ">
      <div className="relative w-full h-full">
        {timeLabels.map((label, index) => (
          <div
            key={index}
            className="absolute top-0 text-xs text-gray-700 dark:text-gray-300"
            style={{
              left: `${getPositionPercentage(label, timelineStart, totalDuration)}%`,
            }}
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

// converts a timestamp to a date and handles undefined values
function convertTimestampToDate(timestamp: Timestamp | undefined): Date {
  return timestamp ? timestampDate(timestamp) : new Date();
}

// calculates the last time if the job is not successful so we can give the timeline an end date
// TODO: this should be revisited
function getCloseOrErrorOrCancelDate(task: JobRunEvent): Date {
  const errorTask = task.tasks.find((item) => item.error);
  const errorTime = errorTask ? errorTask.eventTime : undefined;
  const cancelTime = task.tasks.find(
    (t) =>
      t.type === 'ActivityTaskCancelRequested' ||
      t.type === 'ActivityTaskCanceled' ||
      t.type === 'ActivityTaskTimedOut' ||
      t.type === 'ActivityTaskTerminated' ||
      // t.type === 'ChildWorkflowExecutionCanceled' || // has issues with close with new?
      t.type === 'ChildWorkflowExecutionTerminated' ||
      t.type === 'ChildWorkflowExecutionTimedOut'
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
    <div className="text-white">{isSyncActivity(task) && schemaTable}</div>
  );
}

function isSyncActivity(task: JobRunEvent): boolean {
  return task.metadata?.metadata.case == 'syncMetadata';
}

interface ActivityLabelProps {
  task: JobRunEvent;
  status: RunStatus;
}

// generates the activity label that we see on the left hand activity column
function ActivityLabel({ task, status }: ActivityLabelProps) {
  return (
    <div className="flex flex-row items-center gap-2 overflow-hidden">
      {task.id.toString()}.
      <TruncatedText text={task.type} />
      <ActivityStatus status={status} />
    </div>
  );
}

// generates the activity status icon that we see on the left hand activity column
function ActivityStatus({ status }: { status: RunStatus }) {
  switch (status) {
    case 'completed':
      return <CheckCircledIcon className="text-green-500" />;
    case 'failed':
      return <CrossCircledIcon className="text-red-500" />;
    case 'canceled':
      return <MinusCircledIcon className="text-yellow-500" />;
    case 'running':
      return <Spinner />;
    case 'terminated':
      return <CrossCircledIcon className="text-gray-500" />;
    default:
      return null;
  }
}

interface StatusFilterProps {
  selectedStatuses: RunStatus[];
  onStatusChange: (status: RunStatus, checked: boolean) => void;
}

// would be nice to replace with a multi-select so you don't have to open/close it everytime you want to make a change
function StatusFilter({ selectedStatuses, onStatusChange }: StatusFilterProps) {
  const uniqueSelectedStatuses = new Set(selectedStatuses);

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
          checked={uniqueSelectedStatuses.has('running')}
          onCheckedChange={(checked) => onStatusChange('running', checked)}
        >
          Running
        </DropdownMenuCheckboxItem>
        <DropdownMenuCheckboxItem
          checked={uniqueSelectedStatuses.has('completed')}
          onCheckedChange={(checked) => onStatusChange('completed', checked)}
        >
          Completed
        </DropdownMenuCheckboxItem>
        <DropdownMenuCheckboxItem
          checked={uniqueSelectedStatuses.has('failed')}
          onCheckedChange={(checked) => onStatusChange('failed', checked)}
        >
          Failed
        </DropdownMenuCheckboxItem>
        <DropdownMenuCheckboxItem
          checked={uniqueSelectedStatuses.has('canceled')}
          onCheckedChange={(checked) => onStatusChange('canceled', checked)}
        >
          Canceled
        </DropdownMenuCheckboxItem>
        <DropdownMenuCheckboxItem
          checked={uniqueSelectedStatuses.has('terminated')}
          onCheckedChange={(checked) => onStatusChange('terminated', checked)}
        >
          Terminated
        </DropdownMenuCheckboxItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function formatFullDate(date: Timestamp | Date | undefined) {
  if (!date) return 'N/A';

  if (isMessage(date, TimestampSchema)) {
    return format(convertTimestampToDate(date), 'MM/dd/yyyy HH:mm:ss:SSS');
  }

  if (date instanceof Date) {
    return format(date, 'MM/dd/yyyy HH:mm:ss:SSS');
  }
}

function formatDate(date: Date): string {
  return format(date, 'MM/dd/yyyy');
}

function formatTime(date: Date): string {
  return format(date, 'HH:mm:ss:SSS');
}

function formatTaskDuration(s: Timestamp | undefined, end: Date | undefined) {
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
}

// handles getting the activity statuses by remaping the ActivityStatuses since (i think) we don't want to show all of them per activity
// the event types in the types field are just the stringified Temporal types
function getTaskStatus(
  task: JobRunEvent,
  jobStatus: JobRunStatus | undefined
): RunStatus {
  let isCompleted = false;
  let isFailed = false;
  let isCanceled = false;

  for (const t of task.tasks) {
    switch (t.type) {
      case 'ActivityTaskCompleted':
      case 'ChildWorkflowExecutionCompleted':
        isCompleted = true;
        break;
      case 'ActivityTaskFailed':
      case 'ActivityTaskTimedOut':
      case 'ActivityTaskCanceled':
      case 'ActivityTaskTerminated':
      case 'ChildWorkflowExecutionFailed':
      case 'ChildWorkflowExecutionTimedOut':
      case 'ChildWorkflowExecutionTerminated':
      case 'StartChildWorkflowExecutionFailed':
        isFailed = true;
        break;
      case 'ActivityTaskCancelRequested':
        isCanceled = true;
        break;
      case 'ActivityTaskStarted':
      case 'ActivityTaskScheduled':
      case 'StartChildWorkflowExecutionInitiated':
        break;
    }

    if (t.error) {
      isFailed = true;
    }

    if (isCompleted) break;
  }

  if (isCompleted) return 'completed';
  if (isFailed) return 'failed';

  const isJobTerminated = jobStatus === JobRunStatus.TERMINATED;
  if (isJobTerminated) return 'terminated';
  if (isCanceled) return 'canceled';

  return 'running';
}

// calculates where in the timeline axis something should be relative to the total duration
// also dictates how far right the timeline goes, reduce if you want the timeline shorter or length otherwise
function getPositionPercentage(
  time: Date,
  timelineStart: Date,
  totalDuration: number
) {
  return ((time.getTime() - timelineStart.getTime()) / totalDuration) * 92;
}

interface ExpandedRowProps {
  toggleExpandedRowBody: (val: string) => void;
  isExpanded: boolean;
  isLastItem: boolean;
  task: JobRunEvent;
}

function ExpandedRow(props: ExpandedRowProps): ReactElement<any> {
  const { toggleExpandedRowBody, isExpanded, isLastItem, task } = props;

  return (
    <React.Fragment key={task.id}>
      <div
        className={cn(
          'px-2 h-10 items-center flex cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 '
        )}
        onClick={() => toggleExpandedRowBody(String(task.id))}
      >
        {isExpanded ? (
          <ChevronUpIcon className="ml-auto" />
        ) : (
          <ChevronDownIcon className="ml-auto" />
        )}
      </div>
      {isExpanded && (
        <div
          className={cn(
            'bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700',
            isLastItem && 'border-0'
          )}
        >
          <ExpandedRowBody task={task} />
        </div>
      )}
    </React.Fragment>
  );
}

interface ExpandedRowBodyProps {
  task: JobRunEvent;
}

function ExpandedRowBody(props: ExpandedRowBodyProps): ReactElement<any> {
  const { task } = props;
  const getLabel = (type: string) => {
    switch (type) {
      case 'ActivityTaskScheduled':
      case 'StartChildWorkflowExecutionInitiated':
        return 'Scheduled';
      case 'ActivityTaskStarted':
      case 'ChildWorkflowExecutionStarted':
        return 'Started';
      case 'ActivityTaskCompleted':
      case 'ChildWorkflowExecutionCompleted':
        return 'Completed';
      case 'ActivityTaskFailed':
      case 'ChildWorkflowExecutionFailed':
      case 'StartChildWorkflowExecutionFailed':
        return 'Failed';
      case 'ActivityTaskTimedOut':
      case 'ChildWorkflowExecutionTimedOut':
        return 'Timed Out';
      case 'ActivityTaskCancelRequested':
        return 'Cancel Requested';
      case 'ActivityTaskTerminated':
      case 'ChildWorkflowExecutionTerminated':
        return 'Terminated';
      default:
        return type;
    }
  };

  return (
    <div className="flex flex-col w-full h-[124px] p-2  text-sm border-t border-gray-200 dark:border-gray-700 gap-2">
      {task.tasks.map((subtask, index) => (
        <div key={subtask.id} className="flex flex-row items-center py-1 gap-2">
          <div className="font-semibold w-[90px]">
            {getLabel(subtask.type)}:
          </div>
          <Badge>{formatFullDate(subtask.eventTime)}</Badge>
          <div className="text-gray-500">
            {index > 0
              ? `+${formatTaskDuration(
                  task.tasks[index - 1].eventTime,
                  convertTimestampToDate(subtask.eventTime)
                )}`
              : '-'}
          </div>
          {subtask.error && (
            <div className="text-red-500 ml-2">
              Error: {subtask.error.message}
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
