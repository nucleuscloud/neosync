import { withNeosyncContext } from '@/api-only/neosync-context';
import { IsJobNameAvailableRequest } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  _: RequestContext
): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const name = searchParams.get('name') ?? '';
  const accountId = searchParams.get('accountId') ?? '';
  return withNeosyncContext(async (ctx) => {
    return ctx.jobsClient.isJobNameAvailable(
      new IsJobNameAvailableRequest({
        name,
        accountId,
      })
    );
  })(req);
}
