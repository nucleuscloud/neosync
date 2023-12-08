import { withNeosyncContext } from '@/api-only/neosync-context';
import { RegenerateAccountApiKeyRequest } from '@/neosync-api-client/mgmt/v1alpha1/api_key_pb';
import { NextRequest, NextResponse } from 'next/server';

export async function PUT(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = RegenerateAccountApiKeyRequest.fromJson(await req.json());
    return ctx.apikeyClient.regenerateAccountApiKey(body);
  })(req);
}
