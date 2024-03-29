'use client';
import useSWR, { KeyedMutator, SWRConfiguration } from 'swr';
import { fetcher } from '../fetcher';
import { HookReply } from './types';
import { useGenericErrorToast } from './useGenericErrorToast';
import { useNeosyncUser } from './useNeosyncUser';

export function useNucleusAuthenticatedFetch<T, RawT>(
  fetchUrl: string,
  isReadyCondition: boolean | undefined,
  swrConfig: SWRConfiguration<RawT, Error> | undefined,
  onData: (data: RawT) => T,
  customFetcher?: (url: string) => Promise<RawT>
): HookReply<T>;
export function useNucleusAuthenticatedFetch<T>(
  fetchUrl: string,
  isReadyCondition?: boolean,
  swrConfig?: SWRConfiguration<T, Error>,
  customFetcher?: (url: string) => Promise<T>
): HookReply<T>;
export function useNucleusAuthenticatedFetch<T, RawT = T>(
  fetchUrl: string,
  isReadyCondition: boolean = true,
  swrConfig?: SWRConfiguration<RawT | T, Error>,
  onData?: (data: RawT | undefined) => T,
  customFetcher?: (url: string) => Promise<RawT | T>
): HookReply<RawT | T> {
  const { data: userResp, isLoading: isUserLoading } = useNeosyncUser();
  const isReady = isReadyCondition && !isUserLoading && !!userResp;

  const fetcherToUse = customFetcher ? customFetcher : fetcher;

  const {
    data,
    error,
    mutate,
    isLoading: isDataLoading,
    isValidating,
  } = useSWR<RawT | T, Error>(
    isReady ? fetchUrl : null,
    fetcherToUse,
    swrConfig
  );
  useGenericErrorToast(error);

  // Must include the isReady check, otherwise isLoading is false, but there is no data or error
  const isLoading = !isReady || isDataLoading;

  if (isLoading) {
    return {
      isLoading: true,
      isValidating,
      data: undefined,
      error: undefined,
      mutate: mutate as KeyedMutator<unknown>,
    };
  }

  return {
    data: onData && !error ? onData(data as RawT) : data,
    error,
    isLoading: false,
    isValidating,
    mutate,
  };
}
