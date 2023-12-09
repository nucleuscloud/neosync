import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { IsJobNameAvailableRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const name = searchParams.get('name') ?? '';
  return withNeosyncContext(async (ctx) => {
    return ctx.client.jobs.isJobNameAvailable(
      new IsJobNameAvailableRequest({
        name,
        accountId: params.accountId,
      })
    );
  })(req);
}
