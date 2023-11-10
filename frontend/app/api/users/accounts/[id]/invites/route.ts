import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  GetTeamAccountInvitesRequest,
  InviteUserToTeamAccountRequest,
  RemoveTeamAccountInviteRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.userClient.getTeamAccountInvites(
      new GetTeamAccountInvitesRequest({
        accountId: params.id,
      })
    );
  })(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) =>
    ctx.userClient.inviteUserToTeamAccount(
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
    return ctx.userClient.removeTeamAccountInvite(
      new RemoveTeamAccountInviteRequest({
        id: inviteId,
      })
    );
  })(req);
}
