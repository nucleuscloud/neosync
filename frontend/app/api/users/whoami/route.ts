import { withNeosyncContext } from '@/api-only/neosync-context';
import { SetPersonalAccountRequest, SetUserRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const setUserResp = await ctx.client.userssetUser(new SetUserRequest({}));

    await ctx.client.userssetPersonalAccount(new SetPersonalAccountRequest({}));
    return setUserResp;
  })(req);
}
