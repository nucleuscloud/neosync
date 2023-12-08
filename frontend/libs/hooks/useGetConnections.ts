import { GetConnectionsResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { JsonValue } from '@bufbuild/protobuf';
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
