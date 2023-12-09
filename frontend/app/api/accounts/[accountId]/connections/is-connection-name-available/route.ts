import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { IsConnectionNameAvailableRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const name = searchParams.get('connectionName') ?? '';
  return withNeosyncContext(async (ctx) => {
    return ctx.client.connections.isConnectionNameAvailable(
      new IsConnectionNameAvailableRequest({
        connectionName: name,
        accountId: params.accountId,
      })
    );
  })(req);
}
