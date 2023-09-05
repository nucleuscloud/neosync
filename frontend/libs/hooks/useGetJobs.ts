import { GetJobsResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJobs(): HookReply<GetJobsResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobsResponse,
    JsonValue | GetJobsResponse
  >(`/api/jobs`, undefined, undefined, (data) =>
    data instanceof GetJobsResponse ? data : GetJobsResponse.fromJson(data)
  );
}
