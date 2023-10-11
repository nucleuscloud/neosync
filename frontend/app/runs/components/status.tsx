import { Badge } from '@/components/ui/badge';
import { JobRunStatus } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';

export const JOB_RUN_STATUS = [
  {
    value: JobRunStatus.ERROR,
    badge: <Badge variant="destructive">Error</Badge>,
  },
  {
    value: JobRunStatus.COMPLETE,
    badge: <Badge className="bg-green-600">Complete</Badge>,
  },
  {
    value: JobRunStatus.FAILED,
    badge: <Badge variant="destructive">Failed</Badge>,
  },
  {
    value: JobRunStatus.RUNNING,
    badge: <Badge className="bg-blue-600">Running</Badge>,
  },
  {
    value: JobRunStatus.PENDING,
    badge: <Badge className="bg-purple-600">Running</Badge>,
  },
  {
    value: JobRunStatus.TERMINATED,
    badge: <Badge className="bg-gray-600">Terminated</Badge>,
  },
  {
    value: JobRunStatus.CANCELED,
    badge: <Badge className="bg-gray-600">Canceled</Badge>,
  },
  {
    value: JobRunStatus.UNSPECIFIED,
    badge: <Badge variant="outline">Unknown</Badge>,
  },
];
