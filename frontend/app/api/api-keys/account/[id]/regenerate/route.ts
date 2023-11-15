import { withNeosyncContext } from '@/api-only/neosync-context';
import { RegenerateAccountApiKeyRequest } from '@/neosync-api-client/mgmt/v1alpha1/api_key_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function PUT(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.apikeyClient.regenerateAccountApiKey(
      new RegenerateAccountApiKeyRequest({
        id: params.id,
      })
    );
  })(req);
}
