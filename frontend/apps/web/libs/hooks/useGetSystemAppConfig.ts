import type { SystemAppConfig } from '@/app/config/app-config';
import useSWR, { KeyedMutator } from 'swr';
import type { HookReply } from './types';
import { useGenericErrorToast } from './useGenericErrorToast';

export function useGetSystemAppConfig(): HookReply<SystemAppConfig> {
  const { data, error, mutate, isLoading, isValidating } =
    useSWR(`/api/config`);
  useGenericErrorToast(error);

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
    data: data,
    error,
    isLoading: false,
    isValidating,
    mutate,
  };
}
