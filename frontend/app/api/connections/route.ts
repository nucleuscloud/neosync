import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  CreateConnectionRequest,
  GetConnectionsRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(req: NextRequest): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const accountId = searchParams.get('accountId') ?? '';
  return withNeosyncContext(async (ctx) => {
    return ctx.connectionClient.getConnections(
      new GetConnectionsRequest({
        accountId,
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
