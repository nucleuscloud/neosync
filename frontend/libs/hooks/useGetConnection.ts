'use client';
import { GetConnectionResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { JsonValue } from '@bufbuild/protobuf';
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
