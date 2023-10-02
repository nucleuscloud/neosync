import { withNeosyncContext } from '@/api-only/neosync-context';
import { CreateJobRunRequest } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = CreateJobRunRequest.fromJson(await req.json());
    return ctx.jobsClient.createJobRun(body);
  })(req);
}
