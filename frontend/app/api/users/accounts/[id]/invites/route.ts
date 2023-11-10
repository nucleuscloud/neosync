import { withNeosyncContext } from '@/api-only/neosync-context';
import { GetTeamAccountInvitesRequest } from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
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
