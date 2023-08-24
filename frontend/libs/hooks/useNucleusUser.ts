'use client';
import useSWR, { KeyedMutator } from 'swr';

import { SetUserResponse } from '@/neosync-api-client/mgmt/v1alpha1/user_account_pb';
import { useUser } from '@auth0/nextjs-auth0/client';
import { JsonValue } from '@bufbuild/protobuf';
import { fetcher } from '../fetcher';
import { HookReply } from './types';
import { useGenericErrorToast } from './useGenericErrorToast';

/**
 * Component that returns Nucleus user data.
 * This hook should be called at least once in the app to ensure that the nucleus user record is set.
 */
export function useNucleusUser(suspense?: boolean): HookReply<SetUserResponse> {
  const { user, isLoading: isAuth0UserLoading, error: auth0Error } = useUser();
  const isReady = !isAuth0UserLoading && user && user.sub;
  const {
    data,
    error,
    mutate,
    isLoading: isDataLoading,
    isValidating,
  } = useSWR<JsonValue, Error>(isReady ? `/api/users/whoami` : null, fetcher, {
    suspense: suspense,
  });
  useGenericErrorToast(auth0Error ?? error);

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
    error: auth0Error ? auth0Error : error,
    isLoading: false,
    mutate: mutate as unknown as KeyedMutator<SetUserResponse>,
  };
}
