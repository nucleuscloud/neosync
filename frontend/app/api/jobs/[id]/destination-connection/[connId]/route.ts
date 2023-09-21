import { withNeosyncContext } from '@/api-only/neosync-context';
import { DeleteJobDestinationConnectionRequest } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function DELETE(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.jobsClient.deleteJobDestinationConnection(
      new DeleteJobDestinationConnectionRequest({
        jobId: params.id,
        connectionId: params.connId,
      })
    );
  })(req);
}
