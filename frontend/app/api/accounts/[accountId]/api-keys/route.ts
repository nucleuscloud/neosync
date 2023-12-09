import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import {
  CreateAccountApiKeyRequest,
  GetAccountApiKeysRequest,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.apikeys.getAccountApiKeys(
      new GetAccountApiKeysRequest({
        accountId: params.accountId,
      })
    );
  })(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = CreateAccountApiKeyRequest.fromJson(await req.json());
    return ctx.client.apikeys.createAccountApiKey(body);
  })(req);
}
