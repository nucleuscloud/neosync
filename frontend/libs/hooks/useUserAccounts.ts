import { GetUserAccountsResponse } from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetUserAccounts(
  suspense?: boolean
): HookReply<GetUserAccountsResponse> {
  return useNucleusAuthenticatedFetch<GetUserAccountsResponse, JsonValue>(
    `/api/users/accounts`,
    undefined,
    { suspense },
    (data) => GetUserAccountsResponse.fromJson(data)
  );
}
