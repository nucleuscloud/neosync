import { GetJobRunEventsResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { getRefreshIntervalFn } from '../utils';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

interface GetJobRunEventsOptions {
  refreshIntervalFn?(data: JsonValue): number;
}

export function useGetJobRunEvents(
  runId: string,
  accountId: string,
  opts: GetJobRunEventsOptions = {}
): HookReply<GetJobRunEventsResponse> {
  const { refreshIntervalFn } = opts;
  return useNucleusAuthenticatedFetch<GetJobRunEventsResponse, JsonValue>(
    `/api/runs/${runId}/events?accountId=${accountId}`,
    !!runId || !!accountId,
    {
      refreshInterval: getRefreshIntervalFn(refreshIntervalFn),
    },
    (data) =>
      data instanceof GetJobRunEventsResponse
        ? data
        : GetJobRunEventsResponse.fromJson(data)
  );
}

const TEN_SECONDS = 10 * 1000;

export function refreshEventsWhenEventsIncomplete(data: JsonValue): number {
  const response = GetJobRunEventsResponse.fromJson(data);
  const { isRunComplete } = response;
  return isRunComplete ? 0 : TEN_SECONDS;
}
