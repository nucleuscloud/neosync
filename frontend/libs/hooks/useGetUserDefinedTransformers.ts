import { GetUserDefinedTransformersResponse } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetUserDefinedTransformers(
  accountId: string
): HookReply<GetUserDefinedTransformersResponse> {
  return useNucleusAuthenticatedFetch<
    GetUserDefinedTransformersResponse,
    JsonValue | GetUserDefinedTransformersResponse
  >(
    `/api/transformers/user-defined?accountId=${accountId}`,
    !!accountId,
    undefined,
    (data) =>
      data instanceof GetUserDefinedTransformersResponse
        ? data
        : GetUserDefinedTransformersResponse.fromJson(data)
  );
}
