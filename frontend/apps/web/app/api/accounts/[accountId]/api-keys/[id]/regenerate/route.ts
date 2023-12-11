import { withNeosyncContext } from '@/api-only/neosync-context';
import { RegenerateAccountApiKeyRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function PUT(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = RegenerateAccountApiKeyRequest.fromJson(await req.json());
    return ctx.client.apikeys.regenerateAccountApiKey(body);
  })(req);
}
