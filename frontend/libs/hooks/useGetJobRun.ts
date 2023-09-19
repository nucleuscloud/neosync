import { GetJobRunResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJobRun(runId: string): HookReply<GetJobRunResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobRunResponse,
    JsonValue | GetJobRunResponse
  >(`/api/runs/${runId}`, !!runId, undefined, (data) =>
    data instanceof GetJobRunResponse ? data : GetJobRunResponse.fromJson(data)
  );
}
