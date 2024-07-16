import type { SystemAppConfig } from '@/app/config/app-config';
import { useQuery, UseQueryResult } from '@tanstack/react-query';
import { fetcher } from '../fetcher';

export function useGetSystemAppConfig(): UseQueryResult<SystemAppConfig> {
  return useQuery({
    queryKey: [`/api/config`],
    queryFn: (ctx) => fetcher(ctx.queryKey.join('/')),
  });
}
