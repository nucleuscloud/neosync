import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { GetConnectionSchemaRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.connectiondata.getConnectionSchema(
      new GetConnectionSchemaRequest({
        connectionId: params.id,
      })
    );
  })(req);
}
