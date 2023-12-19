import { JsonValue } from '@bufbuild/protobuf';
import { GetUserAccountsResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetUserAccounts(): HookReply<GetUserAccountsResponse> {
  return useNucleusAuthenticatedFetch<GetUserAccountsResponse, JsonValue>(
    `/api/users/accounts`,
    true,
    undefined,
    (data) => GetUserAccountsResponse.fromJson(data)
  );
}
