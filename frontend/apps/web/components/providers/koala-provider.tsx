'use client';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { useNeosyncUser } from '@/libs/hooks/useNeosyncUser';
import { useSession } from 'next-auth/react';
import Script from 'next/script';
import { ReactElement, useEffect } from 'react';
import { useAccount } from './account-provider';

export function KoalaScriptProvider(): ReactElement {
  const { data: systemAppConfig, isLoading } = useGetSystemAppConfig();

  if (
    isLoading ||
    !systemAppConfig?.koala.enabled ||
    !systemAppConfig.koala.key
  ) {
    return <></>;
  }
  return (
    <Script
      id="koala"
      dangerouslySetInnerHTML={{
        __html: `!function(t){if(window.ko)return;window.ko=[],["identify","track","removeListeners","open","on","off","qualify","ready"].forEach(function(t){ko[t]=function(){var n=[].slice.call(arguments);return n.unshift(t),ko.push(n),ko}});var n=document.createElement("script");n.async=!0,n.setAttribute("src","https://cdn.getkoala.com/v1/${systemAppConfig.koala.key}/sdk.js"),(document.body || document.head).appendChild(n)}();`,
      }}
    />
  );
}

const isBrowser = () => typeof window !== 'undefined';

export function KoalaIdentifier(): ReactElement {
  const { data: systemAppConfig, isLoading: isSystemAppConfigLoading } =
    useGetSystemAppConfig();
  const { data: userData, isLoading: isUserDataLoading } = useNeosyncUser();
  const { data: session } = useSession();
  const { account, isLoading: isAccountLoading } = useAccount();
  const user = session?.user;

  useEffect(() => {
    if (
      !isBrowser() ||
      isUserDataLoading ||
      isAccountLoading ||
      isSystemAppConfigLoading ||
      !systemAppConfig?.koala.enabled ||
      !systemAppConfig.koala.key ||
      !systemAppConfig.isAuthEnabled ||
      !user?.email
    ) {
      return;
    }
    // we only want to set the user id if auth is enabled, otherwise it is always the same
    // so it makes it harder to identify unique posthog sessions when running in un-auth mode.
    const userId = systemAppConfig?.isAuthEnabled
      ? userData?.userId
      : undefined;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (window as any).ko?.identify(user?.email, {
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
    systemAppConfig?.koala.enabled,
    systemAppConfig?.koala.key,
  ]);

  return <></>;
}
