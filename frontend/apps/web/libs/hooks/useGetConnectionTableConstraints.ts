import { JsonValue } from '@bufbuild/protobuf';
import { GetConnectionTableConstraintsResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetConnectionTableConstraints(
  accountId: string,
  connectionId: string
): HookReply<GetConnectionTableConstraintsResponse> {
  return useNucleusAuthenticatedFetch<
    GetConnectionTableConstraintsResponse,
    JsonValue | GetConnectionTableConstraintsResponse
  >(
    `/api/accounts/${accountId}/connections/${connectionId}/table-constraints`,
    !!accountId && !!connectionId,
    undefined,
    (data) =>
      data instanceof GetConnectionTableConstraintsResponse
        ? data
        : GetConnectionTableConstraintsResponse.fromJson(data)
  );
}
