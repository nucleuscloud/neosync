import { GetJobResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetJob(
  accountId: string,
  jobId: string
): HookReply<GetJobResponse> {
  return useNucleusAuthenticatedFetch<
    GetJobResponse,
    JsonValue | GetJobResponse
  >(
    `/api/accounts/${accountId}/jobs/${jobId}`,
    !!accountId && !!jobId,
    undefined,
    (data) =>
      data instanceof GetJobResponse ? data : GetJobResponse.fromJson(data)
  );
}
