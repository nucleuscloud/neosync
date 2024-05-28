import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { GetConnectionTableConstraintsRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.connectiondata.getConnectionTableConstraints(
      new GetConnectionTableConstraintsRequest({
        connectionId: params.id,
      })
    );
  })(req);
}
