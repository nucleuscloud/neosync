import { withNeosyncContext } from '@/api-only/neosync-context';
import { CheckSqlQueryRequest } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.connectionClient.checkSqlQuery(
      new CheckSqlQueryRequest({
        id: params.id,
        query: req.nextUrl.searchParams.get('query') ?? '',
      })
    );
  })(req);
}
