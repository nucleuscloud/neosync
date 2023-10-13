import { GetJobNextRunsResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJobNextRuns(
  jobId: string
): HookReply<GetJobNextRunsResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobNextRunsResponse,
    JsonValue | GetJobNextRunsResponse
  >(`/api/jobs/${jobId}/next-runs`, !!jobId, undefined, (data) =>
    data instanceof GetJobNextRunsResponse
      ? data
      : GetJobNextRunsResponse.fromJson(data)
  );
}
