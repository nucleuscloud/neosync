import { JsonValue } from '@bufbuild/protobuf';
import { GetJobRunsResponse } from '@neosync/sdk';
import { getRefreshIntervalFn } from '../utils';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export type JobRunsAutoRefreshInterval = 'off' | '10s' | '30s' | '1m' | '5m';

interface GetJobRunsOptions {
  refreshIntervalFn?(data: JsonValue): number;
  isPaused?(): boolean;
}

export function useGetJobRuns(
  accountId: string,
  opts: GetJobRunsOptions = {}
): HookReply<GetJobRunsResponse> {
  const { refreshIntervalFn, isPaused } = opts;
  return useNucleusAuthenticatedFetch<GetJobRunsResponse, JsonValue>(
    `/api/accounts/${accountId}/runs`,
    !!accountId,
    {
      refreshInterval: getRefreshIntervalFn(refreshIntervalFn),
      isPaused: isPaused ?? (() => false),
    },
    (data) =>
      data instanceof GetJobRunsResponse
        ? data
        : GetJobRunsResponse.fromJson(data)
  );
}

export function onJobRunsAutoRefreshInterval(
  interval: JobRunsAutoRefreshInterval
): number {
  switch (interval) {
    case 'off':
      return 0;
    case '10s':
      return 10 * 1000;
    case '30s':
      return 30 * 1000;
    case '1m':
      return 1 * 60 * 1000;
    case '5m':
      return 5 * 60 * 1000;
    default:
      return 0;
  }
}

export function onJobRunsPaused(interval: JobRunsAutoRefreshInterval): boolean {
  return interval === 'off';
}
