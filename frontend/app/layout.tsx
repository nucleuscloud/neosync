import { authOptions } from '@/api-only/auth-config';
import '@/app/globals.css';
import SiteFooter from '@/components/SiteFooter';
import SiteHeader from '@/components/SiteHeader';
import AccountProvider from '@/components/providers/account-provider';
import { SessionProvider } from '@/components/providers/session-provider';
import { ThemeProvider } from '@/components/providers/theme-provider';
import { Toaster } from '@/components/ui/toaster';
import { fontSans } from '@/libs/fonts';
import { cn } from '@/libs/utils';
import { Metadata } from 'next';
import { getServerSession } from 'next-auth';

export const metadata: Metadata = {};

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const session = await getServerSession(authOptions);
  return (
    <html lang="en" suppressHydrationWarning>
      <head />
      <body
        className={cn(
          'min-h-screen bg-background font-sans antialiased',
          fontSans.variable
        )}
      >
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <SessionProvider session={session}>
            <AccountProvider>
              <div className="relative flex min-h-screen flex-col">
                <SiteHeader />
                <div className="flex-1 ">{children}</div>
                <SiteFooter />
                <Toaster />
              </div>
            </AccountProvider>
          </SessionProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
