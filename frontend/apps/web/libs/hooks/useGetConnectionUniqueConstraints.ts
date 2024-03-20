import { JsonValue } from '@bufbuild/protobuf';
import { GetConnectionUniqueConstraintsResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetConnectionUniqueConstraints(
  accountId: string,
  connectionId: string
): HookReply<GetConnectionUniqueConstraintsResponse> {
  return useNucleusAuthenticatedFetch<
    GetConnectionUniqueConstraintsResponse,
    JsonValue | GetConnectionUniqueConstraintsResponse
  >(
    `/api/accounts/${accountId}/connections/${connectionId}/unique-constraints`,
    !!accountId && !!connectionId,
    undefined,
    (data) =>
      data instanceof GetConnectionUniqueConstraintsResponse
        ? data
        : GetConnectionUniqueConstraintsResponse.fromJson(data)
  );
}
