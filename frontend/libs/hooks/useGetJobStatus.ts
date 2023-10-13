import { GetJobStatusResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJobStatus(
  jobId: string
): HookReply<GetJobStatusResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobStatusResponse,
    JsonValue | GetJobStatusResponse
  >(`/api/jobs/${jobId}/status`, !!jobId, undefined, (data) =>
    data instanceof GetJobStatusResponse
      ? data
      : GetJobStatusResponse.fromJson(data)
  );
}
