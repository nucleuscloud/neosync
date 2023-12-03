import { GetUserDefinedTransformerByIdResponse } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetUserDefinedTransformersById(
  transformerId: string
): HookReply<GetUserDefinedTransformerByIdResponse> {
  return useNucleusAuthenticatedFetch<
    GetUserDefinedTransformerByIdResponse,
    JsonValue | GetUserDefinedTransformerByIdResponse
  >(
    `/api/transformers/user-defined/${transformerId}`,
    !!transformerId,
    undefined,
    (data) =>
      data instanceof GetUserDefinedTransformerByIdResponse
        ? data
        : GetUserDefinedTransformerByIdResponse.fromJson(data)
  );
}
