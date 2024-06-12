import '@/app/globals.css';
import { KoalaScriptProvider } from '@/components/providers/koala-provider';
import {
  PHProvider,
  PostHogPageview,
} from '@/components/providers/posthog-provider';
import { ThemeProvider } from '@/components/providers/theme-provider';
import { Metadata } from 'next';
import { ReactElement, ReactNode, Suspense } from 'react';
import BaseLayout from '../BaseLayout';

export const metadata: Metadata = {
  title: 'Neosync',
  description: 'Open Source Data Anonymization and Synthetic Data',
  icons: [{ rel: 'icon', url: '/favicon.ico' }],
};

export default async function InviteLayout({
  children,
}: {
  children: ReactNode;
}): Promise<ReactElement> {
  return (
    <html lang="en" suppressHydrationWarning>
      <head />
      <body className="min-h-screen bg-background font-sans antialiased overflow-scroll">
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <>
            <Suspense>
              <KoalaScriptProvider />
            </Suspense>
            <Suspense>
              <PostHogPageview />
            </Suspense>
            <PHProvider>
              <BaseLayout>{children}</BaseLayout>
            </PHProvider>
          </>
        </ThemeProvider>
      </body>
    </html>
  );
}
