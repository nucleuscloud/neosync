import { withNeosyncContext } from '@/api-only/neosync-context';
import { CreateJobRunRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = CreateJobRunRequest.fromJson(await req.json());
    return ctx.client.jobs.createJobRun(body);
  })(req);
}
