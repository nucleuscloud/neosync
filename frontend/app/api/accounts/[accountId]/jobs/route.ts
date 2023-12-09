import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { CreateJobRequest, GetJobsRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.client.jobs.getJobs(
      new GetJobsRequest({
        accountId: params.accountId,
      })
    );
  })(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = CreateJobRequest.fromJson(await req.json());
    return ctx.client.jobs.createJob(body);
  })(req);
}
