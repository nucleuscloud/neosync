import { withNeosyncContext } from '@/api-only/neosync-context';
import { Timestamp } from '@bufbuild/protobuf';
import { GetMetricCountRequest } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(req: NextRequest): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const start = searchParams.get('start') ?? '';
  const end = searchParams.get('end') ?? '';
  const metric = searchParams.get('metric') ?? '';
  const idtype = searchParams.get('idtype') ?? '';
  const identifier = searchParams.get('identifier') ?? '';
  return withNeosyncContext(async (ctx) => {
    const body = new GetMetricCountRequest({
      start: new Timestamp({ seconds: BigInt(parseInt(start, 10)) }),
      end: new Timestamp({ seconds: BigInt(parseInt(end, 10)) }),
      metric: parseInt(metric, 10),
      identifier:
        idtype === 'accountId'
          ? { case: 'accountId', value: identifier }
          : idtype === 'jobId'
            ? { case: 'jobId', value: identifier }
            : idtype === 'runId'
              ? { case: 'runId', value: identifier }
              : undefined,
    });
    return ctx.client.metrics.getMetricCount(body);
  })(req);
}
