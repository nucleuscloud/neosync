'use client';
import { buildCheckConnectionKey } from '@/app/(mgmt)/[account]/connections/util';
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
  let requestBody = {};
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

  const fetcher = (url: string) =>
    fetch(url, {
      method: 'post',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestBody),
      credentials: 'include',
    }).then((res) =>
      res.json().then((body) => {
        if (res.ok) {
          return body;
        }
        if (body.error) {
          throw new Error(body.error);
        }
        if (res.status > 399 && body.message) {
          throw new Error(body.message);
        }
        throw new Error('Unknown error when fetching');
      })
    );

  const a = useNucleusAuthenticatedFetch<
    CheckConnectionConfigResponse,
    JsonValue | CheckConnectionConfigResponse
  >(
    buildCheckConnectionKey(accountId),
    !!accountId && canProceed,
    undefined,
    (data) =>
      data instanceof CheckConnectionConfigResponse
        ? data
        : CheckConnectionConfigResponse.fromJson(data),
    fetcher
  );

  return a;
}
