import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { Code, ConnectError } from '@connectrpc/connect';
import {
  DeleteConnectionRequest,
  GetConnectionRequest,
  UpdateConnectionRequest,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const connection = await ctx.client.connections.getConnection(
      new GetConnectionRequest({
        id: params.id,
      })
    );
    if (connection.connection?.accountId !== params.accountId) {
      throw new ConnectError('resource not found in account', Code.NotFound);
    }
    return connection;
  })(req);
}

export async function PUT(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = UpdateConnectionRequest.fromJson(await req.json());
    return ctx.client.connections.updateConnection(body);
  })(req);
}

export async function DELETE(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.connections.deleteConnection(
      new DeleteConnectionRequest({
        id: params.id,
      })
    );
  })(req);
}
