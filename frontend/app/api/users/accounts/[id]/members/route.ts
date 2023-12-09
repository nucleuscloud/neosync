import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import {
  GetTeamAccountMembersRequest,
  RemoveTeamAccountMemberRequest,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.users.getTeamAccountMembers(
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
    return ctx.client.users.removeTeamAccountMember(
      new RemoveTeamAccountMemberRequest({
        accountId: params.id,
        userId,
      })
    );
  })(req);
}
