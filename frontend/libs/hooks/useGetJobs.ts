import { JsonValue } from '@bufbuild/protobuf';
import { GetJobsResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJobs(accountId: string): HookReply<GetJobsResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobsResponse,
    JsonValue | GetJobsResponse
  >(`/api/accounts/${accountId}/jobs`, !!accountId, undefined, (data) =>
    data instanceof GetJobsResponse ? data : GetJobsResponse.fromJson(data)
  );
}
