import { GetCustomTransformersResponse } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetCustomTransformers(
  accountId: string
): HookReply<GetCustomTransformersResponse> {
  return useNucleusAuthenticatedFetch<
    GetCustomTransformersResponse,
    JsonValue | GetCustomTransformersResponse
  >(
    `/api/transformers/custom?accountId=${accountId}`,
    !!accountId,
    undefined,
    (data) =>
      data instanceof GetCustomTransformersResponse
        ? data
        : GetCustomTransformersResponse.fromJson(data)
  );
}
