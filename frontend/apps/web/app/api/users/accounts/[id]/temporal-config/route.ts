import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import {
  GetAccountTemporalConfigRequest,
  SetAccountTemporalConfigRequest,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) =>
    ctx.client.users.getAccountTemporalConfig(
      new GetAccountTemporalConfigRequest({
        accountId: params.id,
      })
    )
  )(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) =>
    ctx.client.users.setAccountTemporalConfig(
      SetAccountTemporalConfigRequest.fromJson(await req.json())
    )
  )(req);
}

export async function PUT(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) =>
    ctx.client.users.setAccountTemporalConfig(
      SetAccountTemporalConfigRequest.fromJson(await req.json())
    )
  )(req);
}
