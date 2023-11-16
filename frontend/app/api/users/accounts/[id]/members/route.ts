import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  GetTeamAccountMembersRequest,
  RemoveTeamAccountMemberRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.userClient.getTeamAccountMembers(
      new GetTeamAccountMembersRequest({
        accountId: params.id,
      })
    );
  })(req);
}

export async function DELETE(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const userId = searchParams.get('id') ?? '';
  return withNeosyncContext(async (ctx) => {
    return ctx.userClient.removeTeamAccountMember(
      new RemoveTeamAccountMemberRequest({
        accountId: params.id,
        userId,
      })
    );
  })(req);
}
