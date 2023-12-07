import { withNeosyncContext } from '@/api-only/neosync-context';
import { IsTransformerNameAvailableRequest } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const tname = searchParams.get('transformerName') ?? '';
  return withNeosyncContext(async (ctx) => {
    return ctx.transformerClient.isTransformerNameAvailable(
      new IsTransformerNameAvailableRequest({
        transformerName: tname,
        accountId: params.accountId,
      })
    );
  })(req);
}
