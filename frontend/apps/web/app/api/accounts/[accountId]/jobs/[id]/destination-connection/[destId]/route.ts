import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { DeleteJobDestinationConnectionRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function DELETE(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.jobs.deleteJobDestinationConnection(
      new DeleteJobDestinationConnectionRequest({
        destinationId: params.destId,
      })
    );
  })(req);
}
