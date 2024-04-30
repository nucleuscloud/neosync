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
    buildGetConnectionRouteKey(accountId, id),
    !!accountId && !!id,
    undefined,
    (data) =>
      data instanceof GetConnectionResponse
        ? data
        : GetConnectionResponse.fromJson(data)
  );
}

export async function getConnection(
  accountId: string,
  id: string
): Promise<GetConnectionResponse> {
  const res = await fetch(buildGetConnectionRouteKey(accountId, id), {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return GetConnectionResponse.fromJson(await res.json());
}

export function buildGetConnectionRouteKey(
  accountId: string,
  id: string
): string {
  return `/api/accounts/${accountId}/connections/${id}`;
}
