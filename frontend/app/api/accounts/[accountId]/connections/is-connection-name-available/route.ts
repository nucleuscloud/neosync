import { withNeosyncContext } from '@/api-only/neosync-context';
import { IsConnectionNameAvailableRequest } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  _: RequestContext
): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const name = searchParams.get('connectionName') ?? '';
  const accountId = searchParams.get('accountId') ?? '';
  return withNeosyncContext(async (ctx) => {
    return ctx.connectionClient.isConnectionNameAvailable(
      new IsConnectionNameAvailableRequest({
        connectionName: name,
        accountId: accountId,
      })
    );
  })(req);
}
