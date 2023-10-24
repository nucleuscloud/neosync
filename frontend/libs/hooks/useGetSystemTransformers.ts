import { GetSystemTransformersResponse } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetSystemTransformers(): HookReply<GetSystemTransformersResponse> {
  return useNucleusAuthenticatedFetch<
    GetSystemTransformersResponse,
    JsonValue | GetSystemTransformersResponse
  >(`/api/transformers`, undefined, undefined, (data) =>
    data instanceof GetSystemTransformersResponse
      ? data
      : GetSystemTransformersResponse.fromJson(data)
  );
}
