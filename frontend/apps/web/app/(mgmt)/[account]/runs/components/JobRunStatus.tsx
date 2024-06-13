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
  containerClassName?: string;
  badgeClassName?: string;
}

const sharedContainerClassName = 'flex flex-row gap-3 px-2 max-w-28';
const sharedBadgeClassName = 'flex flex-row gap-3';

export default function JobRunStatus(props: Props): ReactElement {
  const { status, containerClassName, badgeClassName } = props;

  switch (status) {
    case JobRunStatusEnum.ERROR:
      return (
        <div className={cn(sharedContainerClassName, containerClassName)}>
          <Badge
            variant="destructive"
            className={cn(sharedBadgeClassName, badgeClassName)}
          >
            <div className="flex flex-row items-center gap-1">
              <ExclamationTriangleIcon className="w-3 h-3" />
              <p>Error</p>
            </div>
          </Badge>
        </div>
      );
    case JobRunStatusEnum.COMPLETE:
      return (
        <div className={cn(sharedContainerClassName, containerClassName)}>
          <Badge
            variant="success"
            className={cn(sharedBadgeClassName, badgeClassName)}
          >
            <CheckIcon className="w-3 h-3" /> <p>Complete</p>
          </Badge>
        </div>
      );
    case JobRunStatusEnum.FAILED:
      return (
        <div className={cn(sharedContainerClassName, containerClassName)}>
          <Badge
            variant="destructive"
            className={cn(sharedBadgeClassName, badgeClassName)}
          >
            <ExclamationTriangleIcon className="w-3 h-3" /> <p>Failed</p>
          </Badge>
        </div>
      );
    case JobRunStatusEnum.RUNNING:
      return (
        <div className={cn(sharedContainerClassName, containerClassName)}>
          <Badge
            className={cn(
              'bg-blue-600 hover:bg-blue-500',
              sharedBadgeClassName,
              badgeClassName
            )}
          >
            <Spinner className="w-3 h-3" /> <p>Running</p>
          </Badge>
        </div>
      );
    case JobRunStatusEnum.PENDING:
      return (
        <div className={cn(sharedContainerClassName, containerClassName)}>
          <Badge
            className={cn(
              'bg-purple-600  hover:bg-purple-500',
              sharedBadgeClassName,
              badgeClassName
            )}
          >
            <Spinner className="w-3 h-3" /> <p>Pending</p>
          </Badge>
        </div>
      );
    case JobRunStatusEnum.TERMINATED:
      return (
        <div className={cn(sharedContainerClassName, containerClassName)}>
          <Badge
            className={cn(
              'bg-gray-600  hover:bg-gray-500',
              sharedBadgeClassName,
              badgeClassName
            )}
          >
            <ExclamationTriangleIcon className="w-3 h-3" /> <p>Terminated</p>
          </Badge>
        </div>
      );
    case JobRunStatusEnum.CANCELED:
      return (
        <div className={cn(sharedContainerClassName, containerClassName)}>
          <Badge
            className={cn(
              'bg-gray-600  hover:bg-gray-500',
              sharedBadgeClassName,
              badgeClassName
            )}
          >
            <ExclamationTriangleIcon className="w-3 h-3" /> <p>Canceled</p>
          </Badge>
        </div>
      );
    case JobRunStatusEnum.TIMED_OUT:
      return (
        <div className={cn(sharedContainerClassName, containerClassName)}>
          <Badge
            className={cn(
              'bg-yellow-600  hover:bg-yellow-500',
              sharedBadgeClassName,
              badgeClassName
            )}
          >
            <ClockIcon className="w-3 h-3" /> <p>Timed Out</p>
          </Badge>
        </div>
      );
    default:
      return (
        <TooltipProvider>
          <Tooltip delayDuration={200}>
            <TooltipTrigger asChild>
              <div className={cn(sharedContainerClassName, containerClassName)}>
                <Badge
                  variant="outline"
                  className={cn(sharedBadgeClassName, badgeClassName)}
                >
                  <QuestionMarkCircledIcon className="w-3 h-3" />
                  <p>Unknown</p>
                </Badge>
              </div>
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
