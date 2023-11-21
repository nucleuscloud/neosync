import {
  GetJobRunResponse,
  JobRunStatus,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { getRefreshIntervalFn } from '../utils';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

interface GetJobRunOptions {
  refreshIntervalFn?(data: JsonValue): number;
}

export function useGetJobRun(
  runId: string,
  accountId: string,
  opts: GetJobRunOptions = {}
): HookReply<GetJobRunResponse> {
  const { refreshIntervalFn } = opts;
  return useNucleusAuthenticatedFetch<GetJobRunResponse, JsonValue>(
    `/api/runs/${runId}?accountId=${accountId}`,
    !!runId || !!accountId,
    {
      refreshInterval: getRefreshIntervalFn(refreshIntervalFn),
    },
    (data) =>
      data instanceof GetJobRunResponse
        ? data
        : GetJobRunResponse.fromJson(data)
  );
}

const TEN_SECONDS = 10 * 1000;

export function refreshWhenJobRunning(data: JsonValue): number {
  const response = GetJobRunResponse.fromJson(data);
  const { jobRun } = response;
  if (!jobRun || !jobRun.status) {
    return 0;
  }
  return shouldRefreshJobRun(jobRun.status) ? TEN_SECONDS : 0;
}

function shouldRefreshJobRun(status?: JobRunStatus): boolean {
  return (
    status === JobRunStatus.RUNNING ||
    status === JobRunStatus.PENDING ||
    status === JobRunStatus.ERROR
  );
}
