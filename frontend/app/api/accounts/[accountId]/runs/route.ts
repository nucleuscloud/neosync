import { withNeosyncContext } from '@/api-only/neosync-context';
import { GetJobRunsRequest } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const jobId = searchParams.get('jobId') ?? '';
  const getJobRunReq = jobId
    ? new GetJobRunsRequest({ id: { value: jobId, case: 'jobId' } })
    : new GetJobRunsRequest({
        id: { value: params.accountId, case: 'accountId' },
      });
  return withNeosyncContext(async (ctx) => {
    return ctx.jobsClient.getJobRuns(getJobRunReq);
  })(req);
}
