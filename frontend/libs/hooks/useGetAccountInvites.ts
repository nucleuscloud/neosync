import { GetTeamAccountInvitesResponse } from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetAccountInvites(
  accountId: string
): HookReply<GetTeamAccountInvitesResponse> {
  return useNucleusAuthenticatedFetch<
    GetTeamAccountInvitesResponse,
    JsonValue | GetTeamAccountInvitesResponse
  >(
    `/api/users/accounts/${accountId}/invites`,
    !!accountId,
    undefined,
    (data) =>
      data instanceof GetTeamAccountInvitesResponse
        ? data
        : GetTeamAccountInvitesResponse.fromJson(data)
  );
}
