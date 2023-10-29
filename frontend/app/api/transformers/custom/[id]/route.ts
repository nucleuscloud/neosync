import { withNeosyncContext } from '@/api-only/neosync-context';
import { GetCustomTransformerByIdRequest } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.transformerClient.getCustomTransformerById(
      new GetCustomTransformerByIdRequest({
        transformerId: params.id,
      })
    );
  })(req);
}
