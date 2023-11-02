import { GetCustomTransformerByIdResponse } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetCustomTransformersById(
  transformerId: string
): HookReply<GetCustomTransformerByIdResponse> {
  return useNucleusAuthenticatedFetch<
    GetCustomTransformerByIdResponse,
    JsonValue | GetCustomTransformerByIdResponse
  >(
    `/api/transformers/custom/${transformerId}`,
    !!transformerId,
    undefined,
    (data) =>
      data instanceof GetCustomTransformerByIdResponse
        ? data
        : GetCustomTransformerByIdResponse.fromJson(data)
  );
}
