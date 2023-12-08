import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  DeleteAccountApiKeyRequest,
  GetAccountApiKeyRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/api_key_pb';
import { RequestContext } from '@/shared';
import { Code, ConnectError } from '@connectrpc/connect';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const apiKey = await ctx.apikeyClient.getAccountApiKey(
      new GetAccountApiKeyRequest({
        id: params.id,
      })
    );
    if (apiKey.apiKey?.accountId !== params.accountId) {
      throw new ConnectError('resource not found in account', Code.NotFound);
    }
    return apiKey;
  })(req);
}

export async function DELETE(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.apikeyClient.deleteAccountApiKey(
      new DeleteAccountApiKeyRequest({
        id: params.id,
      })
    );
  })(req);
}
