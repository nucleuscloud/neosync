import SiteFooter from '@/components/SiteFooter';
import SiteHeader from '@/components/SiteHeader';
import AccountProvider from '@/components/providers/account-provider';
import { SessionProvider } from '@/components/providers/session-provider';
import { Toaster } from '@/components/ui/toaster';
import { ReactElement, ReactNode } from 'react';
import { auth, signIn } from './api/auth/[...nextauth]/auth';
import { getSystemAppConfig } from './api/config/config';

interface Props {
  children: ReactNode;
  // If true, will disable signIn on the server if auth is enabled and there is no session
  disableServerSignin?: boolean;
}

export default async function BaseLayout(props: Props): Promise<ReactElement> {
  const { children, disableServerSignin } = props;
  const systemAppConfig = getSystemAppConfig();
  const session = systemAppConfig.isAuthEnabled ? await auth() : null;
  if (!disableServerSignin && systemAppConfig.isAuthEnabled && !session) {
    await signIn();
  }
  return (
    <SessionProvider session={session}>
      <AccountProvider>
        <div className="relative flex min-h-screen flex-col">
          <SiteHeader />
          <div className="flex-1 container" id="top-level-layout">
            {children}
          </div>
          <SiteFooter />
          <Toaster />
        </div>
      </AccountProvider>
    </SessionProvider>
  );
}
