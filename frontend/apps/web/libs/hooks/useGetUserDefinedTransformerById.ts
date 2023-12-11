import { JsonValue } from '@bufbuild/protobuf';
import { GetUserDefinedTransformerByIdResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetUserDefinedTransformersById(
  accountId: string,
  transformerId: string
): HookReply<GetUserDefinedTransformerByIdResponse> {
  return useNucleusAuthenticatedFetch<
    GetUserDefinedTransformerByIdResponse,
    JsonValue | GetUserDefinedTransformerByIdResponse
  >(
    `/api/accounts/${accountId}/transformers/user-defined/${transformerId}`,
    !!accountId && !!transformerId,
    undefined,
    (data) =>
      data instanceof GetUserDefinedTransformerByIdResponse
        ? data
        : GetUserDefinedTransformerByIdResponse.fromJson(data)
  );
}
