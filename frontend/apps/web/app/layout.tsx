import '@/app/globals.css';
import {
  PHProvider,
  PostHogPageview,
} from '@/components/providers/posthog-provider';
import { ThemeProvider } from '@/components/providers/theme-provider';
import { UnifyScriptProvider } from '@/components/providers/unify-provider';
import { Metadata } from 'next';
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
