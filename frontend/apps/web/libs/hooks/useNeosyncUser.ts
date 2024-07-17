'use client';

import { SetUserResponse } from '@neosync/sdk';
import { useQuery, UseQueryResult } from '@tanstack/react-query';
import { useSession } from 'next-auth/react';
import { fetcher } from '../fetcher';
import { useGetSystemAppConfig } from './useGetSystemAppConfig';

/**
 * Neosync user data.
 * This hook should be called at least once in the app to ensure that the user record is set.
 */
export function useNeosyncUser(): UseQueryResult<SetUserResponse> {
  const { data: systemAppConfigData, isLoading: systemAppConfigLoading } =
    useGetSystemAppConfig();
  const { status } = useSession({
    required: systemAppConfigData?.isAuthEnabled ?? false,
    onUnauthenticated() {
      // override this behavior to prevent routing to the next-auth login page.
      // we can be smarter here and route to the home page..but need to be careful to not do so if already on the / page or the default account page
      console.error('the request is unauthenticated!');
    },
  });
  return useQuery({
    queryKey: ['/api/users/whoami'],
    queryFn: (ctx) => fetcher(ctx.queryKey.join('/')),
    enabled() {
      return (
        !systemAppConfigLoading &&
        isReadyStatus(systemAppConfigData?.isAuthEnabled ?? false, status)
      );
    },
  });
}

function isReadyStatus(isAuthEnabled: boolean, status: string): boolean {
  return isAuthEnabled ? status === 'authenticated' : true;
}
