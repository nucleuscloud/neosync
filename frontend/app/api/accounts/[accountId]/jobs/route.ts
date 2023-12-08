import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  CreateJobRequest,
  GetJobsRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { RequestContext } from '@/shared';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    return ctx.jobsClient.getJobs(
      new GetJobsRequest({
        accountId: params.accountId,
      })
    );
  })(req);
}

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = CreateJobRequest.fromJson(await req.json());
    return ctx.jobsClient.createJob(body);
  })(req);
}
