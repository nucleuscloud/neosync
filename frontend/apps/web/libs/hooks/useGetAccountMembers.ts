import { JsonValue } from '@bufbuild/protobuf';
import { GetTeamAccountMembersResponse } from '@neosync/sdk';
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
