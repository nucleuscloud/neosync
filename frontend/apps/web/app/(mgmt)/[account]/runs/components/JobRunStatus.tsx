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
  Cross2Icon,
  ExclamationTriangleIcon,
  QuestionMarkCircledIcon,
} from '@radix-ui/react-icons';
import { ReactElement } from 'react';

interface Props {
  status?: JobRunStatusEnum;
  className?: string;
}

export default function JobRunStatus(props: Props): ReactElement {
  const { status, className } = props;
  switch (status) {
    case JobRunStatusEnum.ERROR:
      return (
        <Badge variant="destructive" className={cn(className, 'min-w-[95px]')}>
          <div className="flex flex-row items-center gap-1">
            <ExclamationTriangleIcon className="w-3 h-3" />
            <p>Error</p>
          </div>
        </Badge>
      );
    case JobRunStatusEnum.COMPLETE:
      return (
        <Badge variant="success" className={cn(className, 'min-w-[95px]')}>
          <div className="flex flex-row items-center gap-1">
            <CheckIcon className="w-3 h-3" /> <p>Complete</p>
          </div>
        </Badge>
      );
    case JobRunStatusEnum.FAILED:
      return (
        <Badge variant="destructive" className={cn(className, 'min-w-[95px]')}>
          <div className="flex flex-row items-center gap-1">
            <Cross2Icon className="w-3 h-3" /> <p>Failed</p>
          </div>
        </Badge>
      );
    case JobRunStatusEnum.RUNNING:
      return (
        <Badge
          className={cn(
            'bg-blue-600 hover:bg-blue-500 min-w-[95px]',
            className
          )}
        >
          <div className="flex flex-row items-center gap-1">
            <Spinner className="w-3 h-3" /> <p>Running</p>
          </div>
        </Badge>
      );
    case JobRunStatusEnum.PENDING:
      return (
        <Badge
          className={cn(
            'bg-purple-600  hover:bg-purple-500 min-w-[95px]',
            className
          )}
        >
          <div className="flex flex-row items-center gap-1">
            <Spinner className="w-3 h-3" /> <p>Pending</p>
          </div>
        </Badge>
      );
    case JobRunStatusEnum.TERMINATED:
      return (
        <Badge
          className={cn(
            'bg-gray-600  hover:bg-gray-500 min-w-[105px]',
            className
          )}
        >
          <div className="flex flex-row items-center gap-1">
            <Cross2Icon className="w-3 h-3" /> <p>Terminated</p>
          </div>
        </Badge>
      );
    case JobRunStatusEnum.CANCELED:
      return (
        <Badge
          className={cn(
            'bg-gray-600  hover:bg-gray-500 min-w-[95px]',
            className
          )}
        >
          <div className="flex flex-row items-center gap-1">
            <Cross2Icon className="w-3 h-3" /> <p>Canceled</p>
          </div>
        </Badge>
      );
    case JobRunStatusEnum.TIMED_OUT:
      return (
        <Badge
          className={cn(
            'bg-yellow-600  hover:bg-yellow-500 min-w-[120px]',
            className
          )}
        >
          <div className="flex flex-row items-center gap-1">
            <ClockIcon className="w-3 h-3" /> <p>Timed Out</p>
          </div>
        </Badge>
      );
    default:
      return (
        <TooltipProvider>
          <Tooltip delayDuration={500}>
            <TooltipTrigger asChild>
              <div>
                <Badge
                  variant="outline"
                  className={cn(className, 'min-w-[95px]')}
                >
                  <div className="flex flex-row items-center gap-1">
                    <QuestionMarkCircledIcon className="w-3 h-3" />
                    <p>Unknown</p>
                  </div>
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
