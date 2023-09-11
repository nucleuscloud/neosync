import { GetJobRunsResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJobRuns(): HookReply<GetJobRunsResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobRunsResponse,
    JsonValue | GetJobRunsResponse
  >(`/api/runs`, undefined, undefined, (data) =>
    data instanceof GetJobRunsResponse
      ? data
      : GetJobRunsResponse.fromJson(data)
  );
}
