import { withNeosyncContext } from '@/api-only/neosync-context';
import { AcceptTeamAccountInviteRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const token = searchParams.get('token') ?? '';
  return withNeosyncContext(async (ctx) =>
    ctx.client.users.acceptTeamAccountInvite(
      new AcceptTeamAccountInviteRequest({
        token,
      })
    )
  )(req);
}
