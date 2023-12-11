import { JsonValue } from '@bufbuild/protobuf';
import { GetAccountApiKeysResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetAccountApiKeys(
  accountId: string
): HookReply<GetAccountApiKeysResponse> {
  return useNucleusAuthenticatedFetch<
    GetAccountApiKeysResponse,
    JsonValue | GetAccountApiKeysResponse
  >(`/api/accounts/${accountId}/api-keys`, !!accountId, undefined, (data) =>
    data instanceof GetAccountApiKeysResponse
      ? data
      : GetAccountApiKeysResponse.fromJson(data)
  );
}
