import '@/app/globals.css';
import SiteFooter from '@/components/SiteFooter';
import SiteHeader from '@/components/SiteHeader';
import AccountProvider from '@/components/contexts/account-context';
import { ThemeProvider } from '@/components/theme-provider';
import { Toaster } from '@/components/ui/toaster';
import { fontSans } from '@/libs/fonts';
import { cn } from '@/libs/utils';
import { UserProvider } from '@auth0/nextjs-auth0/client';
import { Metadata } from 'next';

export const metadata: Metadata = {};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
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
          <div className="relative flex min-h-screen flex-col">
            <SiteHeader />
            <div className="flex-1 container">
              <UserProvider>
                <AccountProvider>{children}</AccountProvider>
              </UserProvider>
            </div>
            <SiteFooter />
            <Toaster />
          </div>
        </ThemeProvider>
      </body>
    </html>
  );
}
