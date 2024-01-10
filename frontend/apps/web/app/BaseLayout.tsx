import SiteFooter from '@/components/SiteFooter';
import AccountProvider from '@/components/providers/account-provider';
import { PostHogIdentifier } from '@/components/providers/posthog-provider';
import { SessionProvider } from '@/components/providers/session-provider';
import SiteHeader from '@/components/site-header/SiteHeader';
import { Toaster } from '@/components/ui/toaster';
import { isPast, parseISO } from 'date-fns';
import { Session } from 'next-auth/types';
import { ReactElement, ReactNode, Suspense } from 'react';
import { auth, signIn } from './api/auth/[...nextauth]/auth';
import { getSystemAppConfig } from './api/config/config';

interface Props {
  children: ReactNode;
  // If true, will disable signIn on the server if auth is enabled and there is no session
  disableServerSignin?: boolean;
}

export default async function BaseLayout(props: Props): Promise<ReactElement> {
  console.log('process', typeof window);
  const { children, disableServerSignin } = props;
  const systemAppConfig = getSystemAppConfig();
  const session = systemAppConfig.isAuthEnabled ? await auth() : null;
  if (
    !disableServerSignin &&
    systemAppConfig.isAuthEnabled &&
    !isSessionValid(session)
  ) {
    await signIn();
  }
  return (
    <SessionProvider session={session}>
      <AccountProvider>
        <Suspense>
          <PostHogIdentifier systemConfig={systemAppConfig} />
        </Suspense>
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

function isSessionValid(session: Session | null): boolean {
  if (!session) {
    return false;
  }
  const expiryDate = parseISO(session.expires);
  return !isPast(expiryDate);
}
