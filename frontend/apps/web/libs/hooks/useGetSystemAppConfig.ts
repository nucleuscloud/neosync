import type { SystemAppConfig } from '@/app/config/app-config';
import type { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetSystemAppConfig(): HookReply<SystemAppConfig> {
  return useNucleusAuthenticatedFetch<SystemAppConfig>(`/api/config`);
}
