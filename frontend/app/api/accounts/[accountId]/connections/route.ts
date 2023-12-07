import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  CreateConnectionRequest,
  GetConnectionsRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.connectionClient.getConnections(
      new GetConnectionsRequest({
        accountId: params.id,
      })
    );
  })(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = CreateConnectionRequest.fromJson(await req.json());
    return ctx.connectionClient.createConnection(body);
  })(req);
}
