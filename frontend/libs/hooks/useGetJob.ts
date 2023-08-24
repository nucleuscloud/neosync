import { GetJobResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJob(jobId: string): HookReply<GetJobResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobResponse,
    JsonValue | GetJobResponse
  >(`/api/jobs/${jobId}`, !!jobId, undefined, (data) =>
    data instanceof GetJobResponse ? data : GetJobResponse.fromJson(data)
  );
}
