import { JsonValue } from '@bufbuild/protobuf';
import { GetAccountTemporalConfigResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetAccountTemporalConfig(
  accountId: string
): HookReply<GetAccountTemporalConfigResponse> {
  return useNucleusAuthenticatedFetch<
    GetAccountTemporalConfigResponse,
    JsonValue | GetAccountTemporalConfigResponse
  >(
    `/api/users/accounts/${accountId}/temporal-config`,
    !!accountId,
    undefined,
    (data) =>
      data instanceof GetAccountTemporalConfigResponse
        ? data
        : GetAccountTemporalConfigResponse.fromJson(data)
  );
}
