'use client';
import { GetAccountApiKeyResponse } from '@/neosync-api-client/mgmt/v1alpha1/api_key_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetAccountApiKey(
  id: string
): HookReply<GetAccountApiKeyResponse> {
  return useNucleusAuthenticatedFetch<
    GetAccountApiKeyResponse,
    JsonValue | GetAccountApiKeyResponse
  >(`/api/api-keys/account/${id}`, !!id, undefined, (data) =>
    data instanceof GetAccountApiKeyResponse
      ? data
      : GetAccountApiKeyResponse.fromJson(data)
  );
}
