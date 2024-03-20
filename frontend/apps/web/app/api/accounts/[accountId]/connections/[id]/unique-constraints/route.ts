import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { GetConnectionUniqueConstraintsRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.connectiondata.getConnectionUniqueConstraints(
      new GetConnectionUniqueConstraintsRequest({
        connectionId: params.id,
      })
    );
  })(req);
}
