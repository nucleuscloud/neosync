import { withNeosyncContext } from '@/api-only/neosync-context';
import { GetJobRecentRunsRequest } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.jobsClient.getJobRecentRuns(
      new GetJobRecentRunsRequest({
        jobId: params.id,
      })
    );
  })(req);
}
