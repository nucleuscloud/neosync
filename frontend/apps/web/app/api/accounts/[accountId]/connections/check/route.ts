import { withNeosyncContext } from '@/api-only/neosync-context';
import { CheckConnectionConfigRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = CheckConnectionConfigRequest.fromJson(await req.json());
    return ctx.client.connections.checkConnectionConfig(body);
  })(req);
}
