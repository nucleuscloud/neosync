import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { GetConnectionPrimaryConstraintsRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.connectiondata.getConnectionPrimaryConstraints(
      new GetConnectionPrimaryConstraintsRequest({
        connectionId: params.id,
      })
    );
  })(req);
}
