import { Badge } from '@/components/ui/badge';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/libs/utils';
import { JobRunStatus as JobRunStatusEnum } from '@neosync/sdk';
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
        <Badge variant="destructive" className={cn(className)}>
          Error
        </Badge>
      );
    case JobRunStatusEnum.COMPLETE:
      return <Badge className={cn('bg-green-600', className)}>Complete</Badge>;
    case JobRunStatusEnum.FAILED:
      return (
        <Badge variant="destructive" className={cn(className)}>
          Failed
        </Badge>
      );
    case JobRunStatusEnum.RUNNING:
      return <Badge className={cn('bg-blue-600', className)}>Running</Badge>;
    case JobRunStatusEnum.PENDING:
      return <Badge className={cn('bg-purple-600', className)}>Pending</Badge>;
    case JobRunStatusEnum.TERMINATED:
      return <Badge className={cn('bg-gray-600', className)}>Terminated</Badge>;
    case JobRunStatusEnum.CANCELED:
      return <Badge className={cn('bg-gray-600', className)}>Canceled</Badge>;
    default:
      return (
        <TooltipProvider>
          <Tooltip delayDuration={500}>
            <TooltipTrigger asChild>
              <div>
                <Badge variant="outline" className={cn(className)}>
                  Unknown
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
