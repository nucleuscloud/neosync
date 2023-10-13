import { GetJobRecentRunsResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJobRecentRuns(
  jobId: string
): HookReply<GetJobRecentRunsResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobRecentRunsResponse,
    JsonValue | GetJobRecentRunsResponse
  >(`/api/jobs/${jobId}/recent-runs`, !!jobId, undefined, (data) =>
    data instanceof GetJobRecentRunsResponse
      ? data
      : GetJobRecentRunsResponse.fromJson(data)
  );
}
