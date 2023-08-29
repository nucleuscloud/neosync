import { GetTransformersResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetTransformers(): HookReply<GetTransformersResponse> {
  return useNucleusAuthenticatedFetch<
    GetTransformersResponse,
    JsonValue | GetTransformersResponse
  >(`/api/transformers`, undefined, undefined, (data) =>
    data instanceof GetTransformersResponse
      ? data
      : GetTransformersResponse.fromJson(data)
  );
}
