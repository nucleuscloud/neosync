import { withNeosyncContext } from '@/api-only/neosync-context';
import { SetJobSourceSqlConnectionSubsetsRequest } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { NextRequest, NextResponse } from 'next/server';

export async function PUT(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = SetJobSourceSqlConnectionSubsetsRequest.fromJson(
      await req.json()
    );
    return ctx.jobsClient.setJobSourceSqlConnectionSubsets(body);
  })(req);
}
