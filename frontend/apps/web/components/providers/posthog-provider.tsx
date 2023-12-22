'use client';
import { PosthogConfig, SystemAppConfig } from '@/app/config/app-config';
import { useNeosyncUser } from '@/libs/hooks/useNeosyncUser';
import { usePathname, useSearchParams } from 'next/navigation';
import posthog from 'posthog-js';
import { PostHogProvider, usePostHog } from 'posthog-js/react';
import { ReactElement, ReactNode, useEffect } from 'react';
import { useAccount } from './account-provider';

interface PosthogPageviewProps {
  config: PosthogConfig;
}

// Enables posthog, as well as turns on pageview tracking.
export function PostHogPageview(props: PosthogPageviewProps): JSX.Element {
  const { config } = props;
  const pathname = usePathname();
  const searchParams = useSearchParams();

  useEffect(() => {
    if (typeof window !== 'undefined' && config.enabled && config.key) {
      posthog.init(config.key, {
        api_host: config.host,
        capture_pageview: false, // Disable automatic pageview capture, as we capture manually
      });
    }
  }, []);

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

interface Props {
  systemConfig: SystemAppConfig;
}

// Handles setting global user data for the user so that it doesn't have to be set on every capture call.
export function PostHogIdentifier(props: Props): ReactElement {
  const { systemConfig } = props;
  const { data: userData, isLoading: isUserDataLoading } = useNeosyncUser();
  const { account, isLoading: isAccountLoading } = useAccount();
  const posthog = usePostHog();

  useEffect(() => {
    if (
      isUserDataLoading ||
      isAccountLoading ||
      !account?.name ||
      !account?.id ||
      !userData?.userId
    ) {
      return;
    }
    // we only want to set the user id if auth is enabled, otherwise it is always the same
    // so it makes it harder to identify unique posthog sessions when running in un-auth mode.
    const userId = systemConfig.isAuthEnabled ? userData.userId : undefined;
    posthog.identify(userId, {
      accountName: account.name,
      accountId: account.id,
    });
  }, [
    isUserDataLoading,
    isAccountLoading,
    account?.id,
    account?.name,
    userData?.userId,
    systemConfig?.isAuthEnabled,
  ]);
  return <></>;
}
