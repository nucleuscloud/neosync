import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  DeleteJobRequest,
  GetJobRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.jobsClient.getJob(
      new GetJobRequest({
        id: params.id,
      })
    );
  })(req);
}

// export async function PUT(req: NextRequest): Promise<NextResponse> {
//   return withNeosyncContext(async (ctx) => {
//     const body = UpdateConnectionRequest.fromJson(await req.json());
//     return ctx.jobsClient.updateConnection(body);
//   })(req);
// }

export async function DELETE(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.jobsClient.deleteJob(
      new DeleteJobRequest({
        id: params.id,
      })
    );
  })(req);
}
