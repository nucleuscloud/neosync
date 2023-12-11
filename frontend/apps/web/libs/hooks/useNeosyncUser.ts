'use client';
import useSWR, { KeyedMutator } from 'swr';

import { isAuthEnabled } from '@/api-only/auth-config';
import { JsonValue } from '@bufbuild/protobuf';
import { SetUserResponse } from '@neosync/sdk';
import { useSession } from 'next-auth/react';
import { fetcher } from '../fetcher';
import { HookReply } from './types';
import { useGenericErrorToast } from './useGenericErrorToast';

/**
 * Neosync user data.
 * This hook should be called at least once in the app to ensure that the user record is set.
 */
export function useNeosyncUser(): HookReply<SetUserResponse> {
  const { status } = useSession();
  const isReady = isReadyStatus(status);
  const {
    data,
    error,
    mutate,
    isLoading: isDataLoading,
    isValidating,
  } = useSWR<JsonValue, Error>(
    isReady ? `/api/users/whoami` : null,
    fetcher,
    {}
  );
  useGenericErrorToast(error);

  const isLoading = !isReady || isDataLoading;
  if (isLoading) {
    return {
      data: undefined,
      isValidating,
      error: undefined,
      isLoading: true,
      mutate: mutate as KeyedMutator<unknown>,
    };
  }
  return {
    data: data ? SetUserResponse.fromJson(data) : undefined,
    isValidating,
    error: error,
    isLoading: false,
    mutate: mutate as unknown as KeyedMutator<SetUserResponse>,
  };
}

function isReadyStatus(status: string): boolean {
  if (!isAuthEnabled()) {
    return true;
  }
  return status === 'authenticated';
}
