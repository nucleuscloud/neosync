import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { DatabaseColumn, GetConnectionSchemaRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const resp = await ctx.client.connectiondata.getConnectionSchema(
      new GetConnectionSchemaRequest({
        connectionId: params.id,
      })
    );
    const map: Record<string, DatabaseColumn[]> = {};
    resp.schemas.forEach((dbcol) => {
      const key = `${dbcol.schema}.${dbcol.table}`;
      const cols = map[key];
      if (!cols) {
        map[key] = [dbcol];
      } else {
        cols.push(dbcol);
      }
    });
    return { schemaMap: map };
  })(req);
}
