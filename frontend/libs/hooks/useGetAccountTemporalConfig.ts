import { GetAccountTemporalConfigResponse } from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetAccountTemporalConfig(): HookReply<GetAccountTemporalConfigResponse> {
  return useNucleusAuthenticatedFetch<
    GetAccountTemporalConfigResponse,
    JsonValue | GetAccountTemporalConfigResponse
  >(`/api/users/accounts/temporal-config`, undefined, undefined, (data) =>
    data instanceof GetAccountTemporalConfigResponse
      ? data
      : GetAccountTemporalConfigResponse.fromJson(data)
  );
}
