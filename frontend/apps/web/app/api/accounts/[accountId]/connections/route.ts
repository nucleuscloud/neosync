import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { CreateConnectionRequest, GetConnectionsRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.connections.getConnections(
      new GetConnectionsRequest({
        accountId: params.accountId,
      })
    );
  })(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = CreateConnectionRequest.fromJson(await req.json());
    return ctx.client.connections.createConnection(body);
  })(req);
}
