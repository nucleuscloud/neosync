import { JsonValue } from '@bufbuild/protobuf';
import { GetJobRunLogsStreamResponse, LogLevel } from '@neosync/sdk';
import { getRefreshIntervalFn } from '../utils';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

interface GetJobRunLogsOptions {
  refreshIntervalFn?(data: JsonValue): number;
}

export function useGetJobRunLogs(
  runId: string,
  accountId: string,
  loglevel: LogLevel,
  opts: GetJobRunLogsOptions = {}
): HookReply<GetJobRunLogsStreamResponse[]> {
  const { refreshIntervalFn } = opts;
  const query = new URLSearchParams({
    loglevel: loglevel.toString(),
  });
  return useNucleusAuthenticatedFetch<GetJobRunLogsStreamResponse[], JsonValue>(
    `/api/accounts/${accountId}/runs/${runId}/logs?${query.toString()}`,
    !!runId || !!accountId,
    {
      refreshInterval: getRefreshIntervalFn(refreshIntervalFn),
    },
    (data) => {
      const dataArr = Array.isArray(data) ? data : [data];
      return dataArr.map((d) =>
        d instanceof GetJobRunLogsStreamResponse
          ? d
          : GetJobRunLogsStreamResponse.fromJson(d)
      );
    }
  );
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
