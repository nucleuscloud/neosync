import '@/app/globals.css';
import {
  PHProvider,
  PostHogPageview,
} from '@/components/providers/posthog-provider';
import { ThemeProvider } from '@/components/providers/theme-provider';
import { fontSans } from '@/libs/fonts';
import { cn } from '@/libs/utils';
import { Metadata } from 'next';
import { ReactElement, ReactNode, Suspense } from 'react';
import BaseLayout from '../BaseLayout';
import { getSystemAppConfig } from '../api/config/config';

export const metadata: Metadata = {
  title: 'Neosync',
  description: 'Open Source Test Data Management',
  icons: [{ rel: 'icon', url: 'favicon.ico' }],
};

export default async function InviteLayout({
  children,
}: {
  children: ReactNode;
}): Promise<ReactElement> {
  const appConfig = getSystemAppConfig();
  return (
    <html lang="en" suppressHydrationWarning>
      <head />
      <body
        className={cn(
          'min-h-screen bg-background font-sans antialiased overflow-scroll',
          fontSans.variable
        )}
      >
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <>
            <Suspense>
              <PostHogPageview config={appConfig.posthog} />
            </Suspense>
            <PHProvider>
              {/*   // Server Signin is disabled for the invite page due to inability to access path or search params on the server
  // Without this, the signin redirect url is set to the root url instead of /invite?token=<token> */}
              <BaseLayout systemAppConfig={appConfig} disableServerSignin>
                {children}
              </BaseLayout>
            </PHProvider>
          </>
        </ThemeProvider>
      </body>
    </html>
  );
}
