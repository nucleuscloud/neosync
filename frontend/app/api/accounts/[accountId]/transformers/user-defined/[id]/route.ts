import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { Code, ConnectError } from '@connectrpc/connect';
import { GetUserDefinedTransformerByIdRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const transformer =
      await ctx.client.transformers.getUserDefinedTransformerById(
        new GetUserDefinedTransformerByIdRequest({
          transformerId: params.id,
        })
      );
    if (transformer.transformer?.accountId !== params.accountId) {
      throw new ConnectError('resource not found in account', Code.NotFound);
    }

    return transformer;
  })(req);
}
