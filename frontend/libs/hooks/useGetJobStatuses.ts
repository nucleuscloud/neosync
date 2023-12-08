import { GetJobStatusesResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJobStatuses(
  accountId: string
): HookReply<GetJobStatusesResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobStatusesResponse,
    JsonValue | GetJobStatusesResponse
  >(
    `/api/accounts/${accountId}/jobs/statuses`,
    !!accountId,
    undefined,
    (data) =>
      data instanceof GetJobStatusesResponse
        ? data
        : GetJobStatusesResponse.fromJson(data)
  );
}
