import { JsonValue } from '@bufbuild/protobuf';
import { GetConnectionsResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetConnections(
  accountId: string
): HookReply<GetConnectionsResponse> {
  return useNucleusAuthenticatedFetch<
    GetConnectionsResponse,
    JsonValue | GetConnectionsResponse
  >(`/api/accounts/${accountId}/connections`, !!accountId, undefined, (data) =>
    data instanceof GetConnectionsResponse
      ? data
      : GetConnectionsResponse.fromJson(data)
  );
}
