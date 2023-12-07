import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  DeleteConnectionRequest,
  GetConnectionRequest,
  UpdateConnectionRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const connection = await ctx.connectionClient.getConnection(
      new GetConnectionRequest({
        id: params.id,
      })
    );
    if (connection.connection?.accountId !== params.accountId) {
      throw new Error('resource not found in account');
    }
    return connection;
  })(req);
}

export async function PUT(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = UpdateConnectionRequest.fromJson(await req.json());
    return ctx.connectionClient.updateConnection(body);
  })(req);
}

export async function DELETE(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.connectionClient.deleteConnection(
      new DeleteConnectionRequest({
        id: params.id,
      })
    );
  })(req);
}
