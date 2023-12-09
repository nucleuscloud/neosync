import { withNeosyncContext } from '@/api-only/neosync-context';
import { SetPersonalAccountRequest, SetUserRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const setUserResp = await ctx.client.users.setUser(new SetUserRequest({}));

    await ctx.client.users.setPersonalAccount(
      new SetPersonalAccountRequest({})
    );
    return setUserResp;
  })(req);
}
