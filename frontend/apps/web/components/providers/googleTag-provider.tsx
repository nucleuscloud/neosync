'use client';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { useNeosyncUser } from '@/libs/hooks/useNeosyncUser';
import { useSession } from 'next-auth/react';
import Script from 'next/script';
import { ReactElement, useEffect } from 'react';
import { useAccount } from './account-provider';

export function GoogleScriptProvider(): ReactElement {
  const { data: systemAppConfig, isLoading } = useGetSystemAppConfig();

  if (
    isLoading ||
    !systemAppConfig?.gtag.enabled ||
    !systemAppConfig.gtag.key
  ) {
    return <></>;
  }
  return (
    <>
      <Script
        src={`https://www.googletagmanager.com/gtag/js?id=${systemAppConfig.gtag.key}`}
        strategy="afterInteractive"
      />
      <Script
        id="google-analytics"
        strategy="afterInteractive"
        dangerouslySetInnerHTML={{
          __html: `
        window.dataLayer = window.dataLayer || [];
        function gtag(){dataLayer.push(arguments);}
        gtag('js', new Date());
        gtag('config', '${systemAppConfig.gtag.key}'}');
      `,
        }}
      />
      <Script
        id="google-ads-conversion"
        strategy="afterInteractive"
        dangerouslySetInnerHTML={{
          __html: `
        gtag('event', 'conversion', {
          'send_to': 'AW-11350558666/9_e_CKSV7uUYEMqPr6Qq'
        });
      `,
        }}
      />
    </>
  );
}

const isBrowser = () => typeof window !== 'undefined';

export function GtagIdentifier(): ReactElement {
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
      !systemAppConfig?.gtag.enabled ||
      !systemAppConfig.gtag.key ||
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
    systemAppConfig?.gtag.enabled,
    systemAppConfig?.gtag.key,
  ]);

  return <></>;
}
