import { withNeosyncContext } from '@/api-only/neosync-context';
import { CancelJobRunRequest } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function PUT(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.jobsClient.cancelJobRun(
      new CancelJobRunRequest({
        jobRunId: params.id,
        accountId: params.accountId,
      })
    );
  })(req);
}
