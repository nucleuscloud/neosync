import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  UpdateJobDestinationConnectionsRequest
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { NextRequest, NextResponse } from 'next/server';

export async function PUT(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = UpdateJobDestinationConnectionsRequest.fromJson(await req.json());
    return ctx.jobsClient.updateJobDestinationConnections(body);
  })(req);
}
