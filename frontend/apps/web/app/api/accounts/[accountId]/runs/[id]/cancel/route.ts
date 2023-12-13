import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { CancelJobRunRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function PUT(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.jobs.cancelJobRun(
      new CancelJobRunRequest({
        jobRunId: params.id,
        accountId: params.accountId,
      })
    );
  })(req);
}
