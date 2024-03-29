'use client';
import { JsonValue } from '@bufbuild/protobuf';
import {
  CheckConnectionConfigResponse,
  PostgresConnectionConfig,
} from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useTestProgressConnection(
  accountId: string,
  data: PostgresConnectionConfig
): HookReply<CheckConnectionConfigResponse> {
  let requestBody;
  let canProceed: boolean = false;
  if (data.connectionConfig.case == 'url') {
    const url = data.connectionConfig.value;
    const tunnel = data.tunnel;
    canProceed = !!url;
    requestBody = { url, tunnel };
  } else if (data.connectionConfig.case == 'connection') {
    const db = data.connectionConfig.value;
    const tunnel = data.tunnel;
    requestBody = { db, tunnel };
    canProceed = !!db.host;
  }

  const fetcher = () =>
    fetch(`/api/accounts/${accountId}/connections/postgres/check`, {
      method: 'post',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    }).then((res) => res.json());

  return useNucleusAuthenticatedFetch<
    CheckConnectionConfigResponse,
    JsonValue | CheckConnectionConfigResponse
  >(
    `/api/accounts/${accountId}/connections/postgres/check`,
    !!accountId && canProceed,
    undefined,
    (data) =>
      data instanceof CheckConnectionConfigResponse
        ? data
        : CheckConnectionConfigResponse.fromJson(data),
    fetcher
  );
}
