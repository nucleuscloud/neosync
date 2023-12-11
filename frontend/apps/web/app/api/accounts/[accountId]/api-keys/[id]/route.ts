import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { Code, ConnectError } from '@connectrpc/connect';
import {
  DeleteAccountApiKeyRequest,
  GetAccountApiKeyRequest,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const apiKey = await ctx.client.apikeys.getAccountApiKey(
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
    return ctx.client.apikeys.deleteAccountApiKey(
      new DeleteAccountApiKeyRequest({
        id: params.id,
      })
    );
  })(req);
}
