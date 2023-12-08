import { GetJobNextRunsResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJobNextRuns(
  accountId: string,
  jobId: string
): HookReply<GetJobNextRunsResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobNextRunsResponse,
    JsonValue | GetJobNextRunsResponse
  >(
    `/api/accounts/${accountId}/jobs/${jobId}/next-runs`,
    !!accountId && !!jobId,
    undefined,
    (data) =>
      data instanceof GetJobNextRunsResponse
        ? data
        : GetJobNextRunsResponse.fromJson(data)
  );
}
