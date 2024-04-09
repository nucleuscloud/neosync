import { JsonValue } from '@bufbuild/protobuf';
import { GetJobRunLogsStreamResponse } from '@neosync/sdk';
import { getRefreshIntervalFn } from '../utils';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

interface GetJobRunLogsOptions {
  refreshIntervalFn?(data: JsonValue): number;
}

export function useGetJobRunLogs(
  runId: string,
  accountId: string,
  opts: GetJobRunLogsOptions = {}
): HookReply<GetJobRunLogsStreamResponse[]> {
  const { refreshIntervalFn } = opts;
  return useNucleusAuthenticatedFetch<GetJobRunLogsStreamResponse[], JsonValue>(
    `/api/accounts/${accountId}/runs/${runId}/logs`,
    !!runId || !!accountId,
    {
      refreshInterval: getRefreshIntervalFn(refreshIntervalFn),
    },
    (data) => {
      if (Array.isArray(data)) {
        return data.map((d) =>
          d instanceof GetJobRunLogsStreamResponse
            ? d
            : GetJobRunLogsStreamResponse.fromJson(d)
        );
      }
      return data instanceof GetJobRunLogsStreamResponse
        ? [data]
        : [GetJobRunLogsStreamResponse.fromJson(data)];
    }
  );
}

const TEN_SECONDS = 5 * 1000;

export function refreshLogsWhenRunNotComplete(data: string[]): number {
  return data.some(
    (l) => l.includes('context canceled') || l.includes('workflow completed')
  )
    ? 0
    : TEN_SECONDS;
}
