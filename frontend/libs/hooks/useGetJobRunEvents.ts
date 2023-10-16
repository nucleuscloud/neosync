import {
  GetJobRunEventsResponse,
  JobRunStatus,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { getRefreshIntervalFn } from '../utils';
import { HookReply } from './types';
import { shouldRefreshJobRun } from './useGetJobRun';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

interface GetJobRunEventsOptions {
  refreshIntervalFn?(data: JsonValue): number;
}

export function useGetJobRunEvents(
  runId: string,
  opts: GetJobRunEventsOptions = {}
): HookReply<GetJobRunEventsResponse> {
  const { refreshIntervalFn } = opts;
  return useNucleusAuthenticatedFetch<GetJobRunEventsResponse, JsonValue>(
    `/api/runs/${runId}/events`,
    !!runId,
    {
      refreshInterval: getRefreshIntervalFn(refreshIntervalFn),
    },
    (data) =>
      data instanceof GetJobRunEventsResponse
        ? data
        : GetJobRunEventsResponse.fromJson(data)
  );
}

export function getRefreshEventsWhenJobRunningFn(
  status?: JobRunStatus
): (data: JsonValue) => number {
  return () => refreshEventsWhenJobRunning(status);
}

const TEN_SECONDS = 10 * 1000;

function refreshEventsWhenJobRunning(status?: JobRunStatus): number {
  if (!status) {
    return 0;
  }
  return shouldRefreshJobRun(status) ? TEN_SECONDS : 0;
}
