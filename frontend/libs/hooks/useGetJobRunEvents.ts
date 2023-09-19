import { GetJobRunEventsResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJobRunEvents(
  runId: string
): HookReply<GetJobRunEventsResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobRunEventsResponse,
    JsonValue | GetJobRunEventsResponse
  >(`/api/runs/${runId}/events`, !!runId, undefined, (data) =>
    data instanceof GetJobRunEventsResponse
      ? data
      : GetJobRunEventsResponse.fromJson(data)
  );
}
