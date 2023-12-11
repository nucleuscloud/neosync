'use client';
import { JsonValue } from '@bufbuild/protobuf';
import { GetAccountApiKeyResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetAccountApiKey(
  accountId: string,
  id: string
): HookReply<GetAccountApiKeyResponse> {
  return useNucleusAuthenticatedFetch<
    GetAccountApiKeyResponse,
    JsonValue | GetAccountApiKeyResponse
  >(
    `/api/accounts/${accountId}/api-keys/${id}`,
    !!accountId && !!id,
    undefined,
    (data) =>
      data instanceof GetAccountApiKeyResponse
        ? data
        : GetAccountApiKeyResponse.fromJson(data)
  );
}
