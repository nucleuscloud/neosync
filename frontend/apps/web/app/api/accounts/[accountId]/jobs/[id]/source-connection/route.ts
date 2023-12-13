import { withNeosyncContext } from '@/api-only/neosync-context';
import { UpdateJobSourceConnectionRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function PUT(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = UpdateJobSourceConnectionRequest.fromJson(await req.json());
    return ctx.client.jobs.updateJobSourceConnection(body);
  })(req);
}
