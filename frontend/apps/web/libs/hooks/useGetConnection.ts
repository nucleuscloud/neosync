'use client';
import { JsonValue } from '@bufbuild/protobuf';
import { GetConnectionResponse } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetConnection(
  accountId: string,
  id: string
): HookReply<GetConnectionResponse> {
  return useNucleusAuthenticatedFetch<
    GetConnectionResponse,
    JsonValue | GetConnectionResponse
  >(
    `/api/accounts/${accountId}/connections/${id}`,
    !!accountId && !!id,
    undefined,
    (data) =>
      data instanceof GetConnectionResponse
        ? data
        : GetConnectionResponse.fromJson(data)
  );
}
