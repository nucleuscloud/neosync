import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import {
  GetTeamAccountInvitesRequest,
  InviteUserToTeamAccountRequest,
  RemoveTeamAccountInviteRequest,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.users.getTeamAccountInvites(
      new GetTeamAccountInvitesRequest({
        accountId: params.id,
      })
    );
  })(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) =>
    ctx.client.users.inviteUserToTeamAccount(
      InviteUserToTeamAccountRequest.fromJson(await req.json())
    )
  )(req);
}

export async function DELETE(
  req: NextRequest,
  _: RequestContext
): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const inviteId = searchParams.get('id') ?? '';
  return withNeosyncContext(async (ctx) => {
    return ctx.client.users.removeTeamAccountInvite(
      new RemoveTeamAccountInviteRequest({
        id: inviteId,
      })
    );
  })(req);
}
