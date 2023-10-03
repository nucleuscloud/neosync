import { withNeosyncContext } from '@/api-only/neosync-context';
import { GetTransformersRequest } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(req: NextRequest): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const accountId = searchParams.get('accountId') ?? '';
  return withNeosyncContext(async (ctx) => {
    return ctx.transformersClient.getTransformers(
      new GetTransformersRequest({ accountId })
    );
  })(req);
}
