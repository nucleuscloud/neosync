import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  SetPersonalAccountRequest,
  SetUserRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const setUserResp = await ctx.userClient.setUser(new SetUserRequest({}));

    await ctx.userClient.setPersonalAccount(new SetPersonalAccountRequest({}));
    return setUserResp;
  })(req);
}
