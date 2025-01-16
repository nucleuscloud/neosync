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
          'send_to': ${systemAppConfig.gtag.conversion}
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
