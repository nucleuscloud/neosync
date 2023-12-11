import { withNeosyncContext } from '@/api-only/neosync-context';
import { GetSystemTransformersRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.transformers.getSystemTransformers(
      new GetSystemTransformersRequest({})
    );
  })(req);
}
