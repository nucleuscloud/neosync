import { GetAccountApiKeysResponse } from '@/neosync-api-client/mgmt/v1alpha1/api_key_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetAccountApiKeys(
  accountId: string
): HookReply<GetAccountApiKeysResponse> {
  return useNucleusAuthenticatedFetch<
    GetAccountApiKeysResponse,
    JsonValue | GetAccountApiKeysResponse
  >(
    `/api/api-keys/account?accountId=${accountId}`,
    !!accountId,
    undefined,
    (data) =>
      data instanceof GetAccountApiKeysResponse
        ? data
        : GetAccountApiKeysResponse.fromJson(data)
  );
}
