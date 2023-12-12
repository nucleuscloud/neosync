import { isAuthEnabled } from '@/api-only/auth-config';

export function useGetAuthEnabled(): boolean {
  return isAuthEnabled();
}
