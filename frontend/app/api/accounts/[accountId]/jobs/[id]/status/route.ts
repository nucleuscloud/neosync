import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { GetJobStatusRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.jobs.getJobStatus(
      new GetJobStatusRequest({
        jobId: params.id,
      })
    );
  })(req);
}
