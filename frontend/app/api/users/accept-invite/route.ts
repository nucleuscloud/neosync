import { withNeosyncContext } from '@/api-only/neosync-context';
import { AcceptTeamAccountInviteRequest } from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const token = searchParams.get('token') ?? '';
  return withNeosyncContext(async (ctx) =>
    ctx.userClient.acceptTeamAccountInvite(
      new AcceptTeamAccountInviteRequest({
        token,
      })
    )
  )(req);
}
