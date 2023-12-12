import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { DeleteJobRunRequest, GetJobRunRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.jobs.getJobRun(
      new GetJobRunRequest({
        jobRunId: params.id,
        accountId: params.accountId,
      })
    );
  })(req);
}

export async function DELETE(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.jobs.deleteJobRun(
      new DeleteJobRunRequest({
        jobRunId: params.id,
        accountId: params.accountId,
      })
    );
  })(req);
}
