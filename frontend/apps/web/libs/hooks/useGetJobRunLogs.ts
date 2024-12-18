import { fromJson, isMessage } from '@bufbuild/protobuf';
import {
  GetJobRunLogsStreamResponse,
  GetJobRunLogsStreamResponseSchema,
  LogLevel,
} from '@neosync/sdk';
import { useQuery, UseQueryResult } from '@tanstack/react-query';
import { fetcher } from '../fetcher';

interface GetJobRunLogsOptions {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  refreshIntervalFn?(data: any): number;
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
        ? refreshIntervalFn(query.state.data)
        : 0;
    },
    select(data) {
      const dataArr = Array.isArray(data) ? data : [data];
      return dataArr.map((d) =>
        isMessage(GetJobRunLogsStreamResponseSchema, d)
          ? d
          : fromJson(GetJobRunLogsStreamResponseSchema, d)
      );
    },
    enabled: !!runId && !!accountId && !!loglevel,
  });
}

const TEN_SECONDS = 5 * 1000;

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function refreshLogsWhenRunNotComplete(data: any): number {
  const dataArr = Array.isArray(data) ? data : [data];
  return dataArr.some((d) => {
    const converted: GetJobRunLogsStreamResponse = isMessage(
      GetJobRunLogsStreamResponseSchema,
      d
    )
      ? d
      : fromJson(GetJobRunLogsStreamResponseSchema, d);
    return (
      converted.logLine.includes('context canceled') ||
      converted.logLine.includes('workflow completed')
    );
  })
    ? 0
    : TEN_SECONDS;
}
