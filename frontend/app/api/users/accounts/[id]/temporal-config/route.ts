import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  GetAccountTemporalConfigRequest,
  SetAccountTemporalConfigRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) =>
    ctx.userClient.getAccountTemporalConfig(
      new GetAccountTemporalConfigRequest({
        accountId: params.id,
      })
    )
  )(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) =>
    ctx.userClient.setAccountTemporalConfig(
      SetAccountTemporalConfigRequest.fromJson(await req.json())
    )
  )(req);
}

export async function PUT(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) =>
    ctx.userClient.setAccountTemporalConfig(
      SetAccountTemporalConfigRequest.fromJson(await req.json())
    )
  )(req);
}
