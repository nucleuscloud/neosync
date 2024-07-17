import { JsonValue } from '@bufbuild/protobuf';
import { GetJobRunLogsStreamResponse, LogLevel } from '@neosync/sdk';
import { useQuery, UseQueryResult } from '@tanstack/react-query';
import { fetcher } from '../fetcher';

interface GetJobRunLogsOptions {
  refreshIntervalFn?(data: JsonValue): number;
}

export function useGetJobRunLogs(
  runId: string,
  accountId: string,
  loglevel: LogLevel,
  opts: GetJobRunLogsOptions = {}
): UseQueryResult<GetJobRunLogsStreamResponse[]> {
  const { refreshIntervalFn } = opts;
  const query = new URLSearchParams({
    loglevel: loglevel.toString(),
  });
  return useQuery({
    queryKey: [
      '/api/accounts',
      accountId,
      'runs',
      runId,
      'logs',
      '?',
      query.toString(),
    ],
    queryFn: (ctx) => fetcher(ctx.queryKey.join('/')),
    refetchInterval(query) {
      return query.state.data && refreshIntervalFn
        ? refreshIntervalFn(query.state.data as JsonValue)
        : 0;
    },
    select(data) {
      const dataArr = Array.isArray(data) ? data : [data];
      return dataArr.map((d) =>
        d instanceof GetJobRunLogsStreamResponse
          ? d
          : GetJobRunLogsStreamResponse.fromJson(d)
      );
    },
    enabled: !!runId && !!accountId && !!loglevel,
  });
}

const TEN_SECONDS = 5 * 1000;

export function refreshLogsWhenRunNotComplete(data: JsonValue): number {
  const dataArr = Array.isArray(data) ? data : [data];
  return dataArr.some((d) => {
    const converted =
      d instanceof GetJobRunLogsStreamResponse
        ? d
        : GetJobRunLogsStreamResponse.fromJson(d);
    return (
      converted.logLine.includes('context canceled') ||
      converted.logLine.includes('workflow completed')
    );
  })
    ? 0
    : TEN_SECONDS;
}
