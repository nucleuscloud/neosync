import { Badge } from '@/components/ui/badge';
import { cn } from '@/libs/utils';
import { JobRunStatus as JobRunStatusEnum } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { ReactElement } from 'react';

interface Props {
  status?: JobRunStatusEnum;
  className?: string;
}

export default function JobRunStatus(props: Props): ReactElement {
  const { status, className } = props;
  if (!status) {
    return <Badge variant="outline">Unknown</Badge>;
  }
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
      return <Badge className={cn('bg-purple-600', className)}>Running</Badge>;
    case JobRunStatusEnum.TERMINATED:
      return <Badge className={cn('bg-gray-600', className)}>Terminated</Badge>;
    case JobRunStatusEnum.CANCELED:
      return <Badge className={cn('bg-gray-600', className)}>Canceled</Badge>;
    default:
      return (
        <Badge variant="outline" className={cn(className)}>
          Unknown
        </Badge>
      );
  }
}
