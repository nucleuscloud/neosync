import { withNeosyncContext } from '@/api-only/neosync-context';
import { SetJobSourceSqlConnectionSubsetsRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function PUT(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = SetJobSourceSqlConnectionSubsetsRequest.fromJson(
      await req.json()
    );
    return ctx.client.jobs.setJobSourceSqlConnectionSubsets(body);
  })(req);
}
