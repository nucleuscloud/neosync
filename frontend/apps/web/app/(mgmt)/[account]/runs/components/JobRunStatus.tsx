import Spinner from '@/components/Spinner';
import { Badge } from '@/components/ui/badge';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/libs/utils';
import { JobRunStatus as JobRunStatusEnum } from '@neosync/sdk';
import {
  CheckIcon,
  ClockIcon,
  ExclamationTriangleIcon,
  QuestionMarkCircledIcon,
} from '@radix-ui/react-icons';
import { ReactElement } from 'react';

interface Props {
  status?: JobRunStatusEnum;
  className?: string;
}

const sharedBadgeClassName = 'flex max-w-28 flex-row gap-3 px-2';

export default function JobRunStatus(props: Props): ReactElement {
  const { status, className } = props;

  switch (status) {
    case JobRunStatusEnum.ERROR:
      return (
        <Badge
          variant="destructive"
          className={cn(sharedBadgeClassName, className)}
        >
          <div className="flex flex-row items-center gap-1">
            <ExclamationTriangleIcon className="w-3 h-3" />
            <p>Error</p>
          </div>
        </Badge>
      );
    case JobRunStatusEnum.COMPLETE:
      return (
        <Badge
          variant="success"
          className={cn(sharedBadgeClassName, className)}
        >
          <CheckIcon className="w-3 h-3" /> <p>Complete</p>
        </Badge>
      );
    case JobRunStatusEnum.FAILED:
      return (
        <Badge
          variant="destructive"
          className={cn(sharedBadgeClassName, className)}
        >
          <ExclamationTriangleIcon className="w-3 h-3" /> <p>Failed</p>
        </Badge>
      );
    case JobRunStatusEnum.RUNNING:
      return (
        <Badge
          className={cn(
            'bg-blue-600 hover:bg-blue-500',
            sharedBadgeClassName,
            className
          )}
        >
          <Spinner className="w-3 h-3" /> <p>Running</p>
        </Badge>
      );
    case JobRunStatusEnum.PENDING:
      return (
        <Badge
          className={cn(
            'bg-purple-600  hover:bg-purple-500',
            sharedBadgeClassName,
            className
          )}
        >
          <Spinner className="w-3 h-3" /> <p>Pending</p>
        </Badge>
      );
    case JobRunStatusEnum.TERMINATED:
      return (
        <Badge
          className={cn(
            'bg-gray-600  hover:bg-gray-500',
            sharedBadgeClassName,
            className
          )}
        >
          <ExclamationTriangleIcon className="w-3 h-3" /> <p>Terminated</p>
        </Badge>
      );
    case JobRunStatusEnum.CANCELED:
      return (
        <Badge
          className={cn(
            'bg-gray-600  hover:bg-gray-500',
            sharedBadgeClassName,
            className
          )}
        >
          <ExclamationTriangleIcon className="w-3 h-3" /> <p>Canceled</p>
        </Badge>
      );
    case JobRunStatusEnum.TIMED_OUT:
      return (
        <Badge
          className={cn(
            'bg-yellow-600  hover:bg-yellow-500',
            sharedBadgeClassName,
            className
          )}
        >
          <ClockIcon className="w-3 h-3" /> <p>Timed Out</p>
        </Badge>
      );
    default:
      return (
        <TooltipProvider>
          <Tooltip delayDuration={200}>
            <TooltipTrigger asChild>
              <Badge
                variant="outline"
                className={cn(sharedBadgeClassName, className)}
              >
                <QuestionMarkCircledIcon className="w-3 h-3" />
                <p>Unknown</p>
              </Badge>
            </TooltipTrigger>
            <TooltipContent>
              <p>
                This run is un-started, archived by the system, or has been
                deleted.
              </p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      );
  }
}
