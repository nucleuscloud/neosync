import { GetJobRecentRunsResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJobRecentRuns(
  accountId: string,
  jobId: string
): HookReply<GetJobRecentRunsResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobRecentRunsResponse,
    JsonValue | GetJobRecentRunsResponse
  >(
    `/api/accounts/${accountId}/jobs/${jobId}/recent-runs`,
    !!accountId && !!jobId,
    undefined,
    (data) =>
      data instanceof GetJobRecentRunsResponse
        ? data
        : GetJobRecentRunsResponse.fromJson(data)
  );
}
