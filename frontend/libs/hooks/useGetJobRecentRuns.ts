import { JsonValue } from '@bufbuild/protobuf';
import { GetJobRecentRunsResponse } from '@neosync/sdk';
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
