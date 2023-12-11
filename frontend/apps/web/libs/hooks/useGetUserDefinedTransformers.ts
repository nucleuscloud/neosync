import { JsonValue } from '@bufbuild/protobuf';
import { GetUserDefinedTransformersResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetUserDefinedTransformers(
  accountId: string
): HookReply<GetUserDefinedTransformersResponse> {
  return useNucleusAuthenticatedFetch<
    GetUserDefinedTransformersResponse,
    JsonValue | GetUserDefinedTransformersResponse
  >(
    `/api/accounts/${accountId}/transformers/user-defined`,
    !!accountId,
    undefined,
    (data) =>
      data instanceof GetUserDefinedTransformersResponse
        ? data
        : GetUserDefinedTransformersResponse.fromJson(data)
  );
}
