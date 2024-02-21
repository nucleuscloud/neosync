import { JsonValue } from '@bufbuild/protobuf';
import { GetConnectionPrimaryConstraintsResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetConnectionPrimaryConstraints(
  accountId: string,
  connectionId: string
): HookReply<GetConnectionPrimaryConstraintsResponse> {
  return useNucleusAuthenticatedFetch<
    GetConnectionPrimaryConstraintsResponse,
    JsonValue | GetConnectionPrimaryConstraintsResponse
  >(
    `/api/accounts/${accountId}/connections/${connectionId}/primary-constraints`,
    !!accountId && !!connectionId,
    undefined,
    (data) =>
      data instanceof GetConnectionPrimaryConstraintsResponse
        ? data
        : GetConnectionPrimaryConstraintsResponse.fromJson(data)
  );
}
