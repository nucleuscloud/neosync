import { GetTransformersResponse } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetTransformers(
  accountId: string
): HookReply<GetTransformersResponse> {
  return useNucleusAuthenticatedFetch<
    GetTransformersResponse,
    JsonValue | GetTransformersResponse
  >(
    `/api/transformers?accountId=${accountId}`,
    !!accountId,
    undefined,
    (data) =>
      data instanceof GetTransformersResponse
        ? data
        : GetTransformersResponse.fromJson(data)
  );
}
