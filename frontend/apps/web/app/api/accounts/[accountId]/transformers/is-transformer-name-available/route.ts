import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { IsTransformerNameAvailableRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const tname = searchParams.get('transformerName') ?? '';
  return withNeosyncContext(async (ctx) => {
    return ctx.client.transformers.isTransformerNameAvailable(
      new IsTransformerNameAvailableRequest({
        transformerName: tname,
        accountId: params.accountId,
      })
    );
  })(req);
}
