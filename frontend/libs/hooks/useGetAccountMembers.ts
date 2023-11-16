import { GetTeamAccountMembersResponse } from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetAccountMembers(
  accountId: string
): HookReply<GetTeamAccountMembersResponse> {
  return useNucleusAuthenticatedFetch<
    GetTeamAccountMembersResponse,
    JsonValue | GetTeamAccountMembersResponse
  >(
    `/api/users/accounts/${accountId}/members`,
    !!accountId,
    undefined,
    (data) =>
      data instanceof GetTeamAccountMembersResponse
        ? data
        : GetTeamAccountMembersResponse.fromJson(data)
  );
}
