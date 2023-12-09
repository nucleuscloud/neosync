import { JsonValue } from '@bufbuild/protobuf';
import { GetJobRunsResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJobRunsByJob(
  accountId: string,
  jobId: string
): HookReply<GetJobRunsResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobRunsResponse,
    JsonValue | GetJobRunsResponse
  >(
    `/api/accounts/${accountId}/runs?jobId=${jobId}`,
    !!accountId && !!jobId,
    undefined,
    (data) =>
      data instanceof GetJobRunsResponse
        ? data
        : GetJobRunsResponse.fromJson(data)
  );
}
