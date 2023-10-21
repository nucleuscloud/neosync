import { withNeosyncContext } from '@/api-only/neosync-context';
import { CheckConnectionConfigRequest } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = CheckConnectionConfigRequest.fromJson(await req.json());
    return ctx.connectionClient.checkConnectionConfig(body);
  })(req);
}
