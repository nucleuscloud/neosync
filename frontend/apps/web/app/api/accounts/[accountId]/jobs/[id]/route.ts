import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { Code, ConnectError } from '@connectrpc/connect';
import { DeleteJobRequest, GetJobRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const job = await ctx.client.jobs.getJob(
      new GetJobRequest({
        id: params.id,
      })
    );

    if (job?.job?.accountId !== params.accountId) {
      throw new ConnectError('resource not found in account', Code.NotFound);
    }

    return job;
  })(req);
}

// export async function PUT(req: NextRequest): Promise<NextResponse> {
//   return withNeosyncContext(async (ctx) => {
//     const body = UpdateConnectionRequest.fromJson(await req.json());
//     return ctx.client.jobs.updateConnection(body);
//   })(req);
// }

export async function DELETE(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.jobs.deleteJob(
      new DeleteJobRequest({
        id: params.id,
      })
    );
  })(req);
}
