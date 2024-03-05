import SiteFooter from '@/components/SiteFooter';
import AccountProvider from '@/components/providers/account-provider';
import { PostHogIdentifier } from '@/components/providers/posthog-provider';
import { SessionProvider } from '@/components/providers/session-provider';
import SiteHeader from '@/components/site-header/SiteHeader';
import { Toaster } from '@/components/ui/toaster';
import { ReactElement, ReactNode, Suspense } from 'react';
import Dopt from './DoptP';
import { auth } from './api/auth/[...nextauth]/auth';

interface Props {
  children: ReactNode;
}

export default async function BaseLayout(props: Props): Promise<ReactElement> {
  const { children } = props;
  const session = await auth();
  return (
    <SessionProvider session={session}>
      <AccountProvider>
        <Suspense>
          <PostHogIdentifier />
        </Suspense>
        <div className="relative flex min-h-screen flex-col">
          <SiteHeader />
          <div className="flex-1 container" id="top-level-layout">
            <Dopt>{children}</Dopt>
          </div>
          <SiteFooter />
          <Toaster />
        </div>
      </AccountProvider>
    </SessionProvider>
  );
}
