import { withNeosyncContext } from '@/api-only/neosync-context';
import { GetDailyMetricCountRequest, Date as NeosyncDate } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(req: NextRequest): Promise<NextResponse> {
  const { searchParams } = new URL(req.url);
  const startDay = parseInt(searchParams.get('startDay') ?? '0', 10);
  const startMo = parseInt(searchParams.get('startMo') ?? '0', 10);
  const startYear = parseInt(searchParams.get('startYear') ?? '0', 10);
  const endDay = parseInt(searchParams.get('endDay') ?? '0', 10);
  const endMo = parseInt(searchParams.get('endMo') ?? '0', 10);
  const endYear = parseInt(searchParams.get('endYear') ?? '0', 10);
  const metric = searchParams.get('metric') ?? '';
  const idtype = searchParams.get('idtype') ?? '';
  const identifier = searchParams.get('identifier') ?? '';
  return withNeosyncContext(async (ctx) => {
    const body = new GetDailyMetricCountRequest({
      start: new NeosyncDate({
        day: startDay,
        month: startMo,
        year: startYear,
      }),
      end: new NeosyncDate({ day: endDay, month: endMo, year: endYear }),
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
    return ctx.client.metrics.getDailyMetricCount(body);
  })(req);
}
