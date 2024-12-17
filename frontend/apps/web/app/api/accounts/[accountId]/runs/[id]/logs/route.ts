import { withNeosyncContext } from '@/api-only/neosync-context';
import { RequestContext } from '@/shared';
import { create } from '@bufbuild/protobuf';
import {
  GetJobRunLogsStreamRequestSchema,
  GetJobRunLogsStreamResponse,
  LogLevel,
  LogWindow,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  { params }: RequestContext
): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const { searchParams } = new URL(req.url);
    const loglevel = searchParams.get('loglevel');
    const response = ctx.client.jobs.getJobRunLogsStream(
      create(GetJobRunLogsStreamRequestSchema, {
        jobRunId: params.id,
        accountId: params.accountId,
        window: getWindow('1d'),
        shouldTail: false,
        maxLogLines: BigInt('5000'),
        logLevels: [loglevel ? parseInt(loglevel, 10) : LogLevel.UNSPECIFIED],
      })
    );
    const logs: GetJobRunLogsStreamResponse[] = [];
    for await (const logRes of response) {
      logs.push(logRes);
    }
    return logs;
  })(req);
}

function getWindow(window?: string): LogWindow {
  if (!window) {
    return LogWindow.NO_TIME_UNSPECIFIED;
  }
  if (window === '15m' || window === '15min') {
    return LogWindow.FIFTEEN_MIN;
  }
  if (window === '1h') {
    return LogWindow.ONE_HOUR;
  }
  if (window === '1d') {
    return LogWindow.ONE_DAY;
  }
  return LogWindow.NO_TIME_UNSPECIFIED;
}
