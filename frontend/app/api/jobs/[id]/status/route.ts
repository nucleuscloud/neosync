import { withNeosyncContext } from '@/api-only/neosync-context';
import { GetJobStatusRequest } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.jobsClient.getJobStatus(
      new GetJobStatusRequest({
        jobId: params.id,
      })
    );
  })(req);
}
