import { withNeosyncContext } from '@/api-only/neosync-context';
import { CreateTeamAccountRequest, GetUserAccountsRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.users.getUserAccounts(new GetUserAccountsRequest({}));
  })(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = CreateTeamAccountRequest.fromJson(await req.json());
    return ctx.client.users.createTeamAccount(body);
  })(req);
}
