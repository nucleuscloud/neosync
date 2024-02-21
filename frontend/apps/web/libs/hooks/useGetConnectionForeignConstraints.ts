import { JsonValue } from '@bufbuild/protobuf';
import { GetConnectionForeignConstraintsResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetConnectionForeignConstraints(
  accountId: string,
  connectionId: string
): HookReply<GetConnectionForeignConstraintsResponse> {
  return useNucleusAuthenticatedFetch<
    GetConnectionForeignConstraintsResponse,
    JsonValue | GetConnectionForeignConstraintsResponse
  >(
    `/api/accounts/${accountId}/connections/${connectionId}/foreign-constraints`,
    !!accountId && !!connectionId,
    undefined,
    (data) =>
      data instanceof GetConnectionForeignConstraintsResponse
        ? data
        : GetConnectionForeignConstraintsResponse.fromJson(data)
  );
}
