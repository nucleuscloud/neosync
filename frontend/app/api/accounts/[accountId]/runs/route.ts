import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { GetJobRunsRequest } from '@neosync/sdk';
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
    return ctx.client.jobs.getJobRuns(getJobRunReq);
  })(req);
}
