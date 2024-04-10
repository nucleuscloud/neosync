import { PlainMessage } from '@bufbuild/protobuf';
import { DatabaseColumn } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export type ConnectionSchemaMap = Record<
  string,
  PlainMessage<DatabaseColumn>[]
>;

export interface GetConnectionSchemaMapResponse {
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

export async function getConnectionSchema(
  accountId: string,
  connectionId?: string
): Promise<GetConnectionSchemaMapResponse | undefined> {
  if (!accountId || !connectionId) {
    return;
  }
  const res = await fetch(
    `/api/accounts/${accountId}/connections/${connectionId}/schema/map`,
    {
      method: 'GET',
      headers: {
        'content-type': 'application/json',
      },
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return res.json();
}
