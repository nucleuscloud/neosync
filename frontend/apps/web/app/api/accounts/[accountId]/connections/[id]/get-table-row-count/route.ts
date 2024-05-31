import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { GetTableRowCountRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    console.log(req.nextUrl.searchParams.get('where'));
    return ctx.client.connectiondata.getTableRowCount(
      new GetTableRowCountRequest({
        connectionId: params.id,
        schema: req.nextUrl.searchParams.get('schema') ?? '',
        table: req.nextUrl.searchParams.get('table') ?? '',
        whereClause: req.nextUrl.searchParams.get('where') ?? undefined,
      })
    );
  })(req);
}
