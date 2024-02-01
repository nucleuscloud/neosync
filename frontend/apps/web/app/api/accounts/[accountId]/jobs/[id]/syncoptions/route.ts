import { withNeosyncContext } from '@/api-only/neosync-context';
import { SetJobSyncOptionsRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function PUT(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = SetJobSyncOptionsRequest.fromJson(await req.json());
    return ctx.client.jobs.setJobSyncOptions(body);
  })(req);
}
