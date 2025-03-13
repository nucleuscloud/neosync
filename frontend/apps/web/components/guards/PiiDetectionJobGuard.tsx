import { isPiiDetectJob } from '@/app/(mgmt)/[account]/jobs/[id]/util';
import { useQuery } from '@connectrpc/connect-query';
import { JobService } from '@neosync/sdk';
import { ReactElement } from 'react';

interface Props {
  jobId: string;
  children: ReactElement;
}
export default function PiiDetectionJobGuard(
  props: Props
): ReactElement | null {
  const { jobId, children } = props;
  const { data: jobResp } = useQuery(
    JobService.method.getJob,
    {
      id: jobId,
    },
    {
      enabled: !!jobId,
    }
  );
  const job = jobResp?.job;
  if (!job || !isPiiDetectJob(job)) {
    return null;
  }

  return children;
}
