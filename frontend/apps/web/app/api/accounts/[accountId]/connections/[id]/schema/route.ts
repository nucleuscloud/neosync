import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { DatabaseColumn, GetConnectionSchemaRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

const fakeSchema: DatabaseColumn[] = [];
for (let i = 0; i < 20000; i++) {
  fakeSchema.push(
    new DatabaseColumn({
      schema: `testschema_${i}`,
      table: `testtable_${i}`,
      column: 'id',
      dataType: 'string',
    })
  );
}

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
    resp.schemas.push(...fakeSchema);
    return resp;
  })(req);
}
