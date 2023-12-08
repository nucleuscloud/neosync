import { withNeosyncContext } from '@/api-only/neosync-context';
import { GetUserDefinedTransformerByIdRequest } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { RequestContext } from '@/shared';
import { Code, ConnectError } from '@connectrpc/connect';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const transformer =
      await ctx.transformerClient.getUserDefinedTransformerById(
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
