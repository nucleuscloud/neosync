import { GetJobRunsResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJobRuns(
  accountId: string
): HookReply<GetJobRunsResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobRunsResponse,
    JsonValue | GetJobRunsResponse
  >(`/api/runs?accountId=${accountId}`, !!accountId, undefined, (data) =>
    data instanceof GetJobRunsResponse
      ? data
      : GetJobRunsResponse.fromJson(data)
  );
}
