import { GetConnectionSchemaResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { JsonValue } from '@bufbuild/protobuf';
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
