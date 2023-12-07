import { withNeosyncContext } from '@/api-only/neosync-context';
import { CancelJobRunRequest } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function PUT(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const accountId = searchParams.get('accountId') ?? '';
  return withNeosyncContext(async (ctx) => {
    return ctx.jobsClient.cancelJobRun(
      new CancelJobRunRequest({
        jobRunId: params.id,
        accountId,
      })
    );
  })(req);
}
