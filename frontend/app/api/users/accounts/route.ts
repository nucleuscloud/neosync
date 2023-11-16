import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  CreateTeamAccountRequest,
  GetUserAccountsRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.userClient.getUserAccounts(new GetUserAccountsRequest({}));
  })(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = CreateTeamAccountRequest.fromJson(await req.json());
    return ctx.userClient.createTeamAccount(body);
  })(req);
}
