import { PlainMessage } from '@bufbuild/protobuf';
import { DatabaseColumn } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export type ConnectionSchemaMap = Record<
  string,
  PlainMessage<DatabaseColumn>[]
>;

interface GetConnectionSchemaMapResponse {
  schemaMap: ConnectionSchemaMap;
}

export function useGetConnectionSchemaMap(
  accountId: string,
  connectionId?: string
): HookReply<GetConnectionSchemaMapResponse> {
  return useNucleusAuthenticatedFetch<
    GetConnectionSchemaMapResponse,
    GetConnectionSchemaMapResponse
  >(
    `/api/accounts/${accountId}/connections/${connectionId}/schema/map`,
    !!accountId && !!connectionId,
    undefined,
    (data) => data
  );
}
