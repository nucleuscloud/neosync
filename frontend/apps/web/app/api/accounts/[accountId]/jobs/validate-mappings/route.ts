import { withNeosyncContext } from '@/api-only/neosync-context';
import { ValidateJobMappingsRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = ValidateJobMappingsRequest.fromJson(await req.json());
    return ctx.client.jobs.validateJobMappings(body);
  })(req);
}
