import { withNeosyncContext } from '@/api-only/neosync-context';
import { getSystemAppConfig } from '@/app/api/config/config';
import { RequestContext } from '@/shared';
import {
  Code,
  ConnectError,
  GetAccountTemporalConfigRequest,
  SetAccountTemporalConfigRequest,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  const systemConfig = getSystemAppConfig();
  return withNeosyncContext(async (ctx) => {
    if (systemConfig.isNeosyncCloud) {
      throw new ConnectError('unimplemented', Code.Unimplemented);
    }
    return ctx.client.users.getAccountTemporalConfig(
      new GetAccountTemporalConfigRequest({
        accountId: params.id,
      })
    );
  })(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  const systemConfig = getSystemAppConfig();
  return withNeosyncContext(async (ctx) => {
    if (systemConfig.isNeosyncCloud) {
      throw new ConnectError('unimplemented', Code.Unimplemented);
    }
    return ctx.client.users.setAccountTemporalConfig(
      SetAccountTemporalConfigRequest.fromJson(await req.json())
    );
  })(req);
}

export async function PUT(req: NextRequest): Promise<NextResponse> {
  const systemConfig = getSystemAppConfig();
  return withNeosyncContext(async (ctx) => {
    if (systemConfig.isNeosyncCloud) {
      throw new ConnectError('unimplemented', Code.Unimplemented);
    }
    return ctx.client.users.setAccountTemporalConfig(
      SetAccountTemporalConfigRequest.fromJson(await req.json())
    );
  })(req);
}
