import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  CreateAccountApiKeyRequest,
  GetAccountApiKeysRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/api_key_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.apikeyClient.getAccountApiKeys(
      new GetAccountApiKeysRequest({
        accountId: params.id,
      })
    );
  })(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = CreateAccountApiKeyRequest.fromJson(await req.json());
    return ctx.apikeyClient.createAccountApiKey(body);
  })(req);
}
