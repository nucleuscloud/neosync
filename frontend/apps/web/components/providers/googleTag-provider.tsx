'use client';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import Script from 'next/script';
import { ReactElement } from 'react';

export function GoogleScriptProvider(): ReactElement<any> {
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
