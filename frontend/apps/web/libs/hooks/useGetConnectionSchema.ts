import { JsonValue } from '@bufbuild/protobuf';
import { GetConnectionSchemaResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetConnectionSchema(
  accountId: string,
  connectionId?: string
): HookReply<GetConnectionSchemaResponse> {
  return useNucleusAuthenticatedFetch<
    GetConnectionSchemaResponse,
    JsonValue | GetConnectionSchemaResponse
  >(
    `/api/accounts/${accountId}/connections/${connectionId}/schema`,
    !!accountId && !!connectionId,
    undefined,
    (data) =>
      data instanceof GetConnectionSchemaResponse
        ? data
        : GetConnectionSchemaResponse.fromJson(data)
  );
}
