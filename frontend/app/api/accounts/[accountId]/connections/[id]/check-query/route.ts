import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { CheckSqlQueryRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.connections.checkSqlQuery(
      new CheckSqlQueryRequest({
        id: params.id,
        query: req.nextUrl.searchParams.get('query') ?? '',
      })
    );
  })(req);
}
