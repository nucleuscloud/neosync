import { JsonValue } from '@bufbuild/protobuf';
import { GetJobStatusResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJobStatus(
  accountId: string,
  jobId: string
): HookReply<GetJobStatusResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobStatusResponse,
    JsonValue | GetJobStatusResponse
  >(
    `/api/accounts/${accountId}/jobs/${jobId}/status`,
    !!accountId && !!jobId,
    undefined,
    (data) =>
      data instanceof GetJobStatusResponse
        ? data
        : GetJobStatusResponse.fromJson(data)
  );
}
