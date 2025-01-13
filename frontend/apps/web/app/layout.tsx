import '@/app/globals.css';
import {
  PHProvider,
  PostHogPageview,
} from '@/components/providers/posthog-provider';
import { ThemeProvider } from '@/components/providers/theme-provider';
import { UnifyScriptProvider } from '@/components/providers/unify-provider';
import { Metadata } from 'next';
import Script from 'next/script';
import { ReactElement, Suspense } from 'react';
import BaseLayout from './BaseLayout';

export const metadata: Metadata = {
  title: 'Neosync',
  description: 'Open Source Data Anonymization and Synthetic Data',
  icons: [{ rel: 'icon', url: '/favicon.ico' }],
};

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode;
}): Promise<ReactElement> {
  return (
    <html lang="en" suppressHydrationWarning>
      <head />
      <body className="min-h-screen bg-background font-sans antialiased overflow-scroll">
        <Script
          src={`https://www.googletagmanager.com/gtag/js?id=G-JNE89GJ6WJ}`}
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
              gtag('config', 'G-JNE89GJ6WJ'}');
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
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <>
            <PHProvider>
              <BaseLayout>
                <>
                  <Suspense>
                    <UnifyScriptProvider />
                  </Suspense>
                  <Suspense>
                    <PostHogPageview />
                  </Suspense>
                  {children}
                </>
              </BaseLayout>
            </PHProvider>
          </>
        </ThemeProvider>
      </body>
    </html>
  );
}
