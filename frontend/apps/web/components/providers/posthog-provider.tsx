'use client';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { useNeosyncUser } from '@/libs/hooks/useNeosyncUser';
import { useSession } from 'next-auth/react';
import { usePathname, useSearchParams } from 'next/navigation';
import posthog from 'posthog-js';
import { PostHogProvider, usePostHog } from 'posthog-js/react';
import { ReactElement, ReactNode, useEffect, type JSX } from 'react';
import { useAccount } from './account-provider';

// Enables posthog, as well as turns on pageview tracking.
export function PostHogPageview(): JSX.Element {
  const { data: systemAppConfig, isLoading: isSystemAppConfigLoading } =
    useGetSystemAppConfig();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  useEffect(() => {
    if (
      typeof window !== 'undefined' &&
      !isSystemAppConfigLoading &&
      systemAppConfig?.posthog?.enabled &&
      systemAppConfig?.posthog?.key
    ) {
      posthog.init(systemAppConfig.posthog.key, {
        api_host: systemAppConfig.posthog.host,
        capture_pageview: false, // Disable automatic pageview capture, as we capture manually
      });
    }
  }, [
    systemAppConfig?.posthog?.enabled,
    systemAppConfig?.posthog?.key,
    isSystemAppConfigLoading,
  ]);

  useEffect(() => {
    if (pathname) {
      let url = window.origin + pathname;
      if (searchParams && searchParams.toString()) {
        url = url + `?${searchParams.toString()}`;
      }
      posthog.capture('$pageview', {
        $current_url: url,
      });
    }
  }, [pathname, searchParams]);

  return <></>;
}

interface PHProps {
  children: ReactNode;
}

// Enables Posthog to be used througout the app via the hook.
export function PHProvider({ children }: PHProps) {
  return <PostHogProvider client={posthog}>{children}</PostHogProvider>;
}

// Handles setting global user data for the user so that it doesn't have to be set on every capture call.
export function PostHogIdentifier(): ReactElement<any> {
  const { data: systemAppConfig, isLoading: isSystemAppConfigLoading } =
    useGetSystemAppConfig();
  const { data: userData, isLoading: isUserDataLoading } = useNeosyncUser();
  const { data: session } = useSession();
  const { account, isLoading: isAccountLoading } = useAccount();
  const posthog = usePostHog();
  const user = session?.user;

  useEffect(() => {
    if (isUserDataLoading || isAccountLoading || isSystemAppConfigLoading) {
      return;
    }
    // we only want to set the user id if auth is enabled, otherwise it is always the same
    // so it makes it harder to identify unique posthog sessions when running in un-auth mode.
    const userId = systemAppConfig?.isAuthEnabled
      ? userData?.userId
      : undefined;
    posthog.identify(userId, {
      accountName: account?.name,
      accountId: account?.id,
      email: user?.email,
      name: user?.name,
      neosyncCloud: systemAppConfig?.isNeosyncCloud ?? false,
      userId,
    });
  }, [
    isUserDataLoading,
    isAccountLoading,
    isSystemAppConfigLoading,
    account?.id,
    account?.name,
    userData?.userId,
    systemAppConfig?.isAuthEnabled,
    systemAppConfig?.isNeosyncCloud,
    user?.email,
    user?.name,
  ]);
  return <></>;
}
