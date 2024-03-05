import SiteFooter from '@/components/SiteFooter';
import AccountProvider from '@/components/providers/account-provider';
import { PostHogIdentifier } from '@/components/providers/posthog-provider';
import { SessionProvider } from '@/components/providers/session-provider';
import SiteHeader from '@/components/site-header/SiteHeader';
import { Toaster } from '@/components/ui/toaster';
import { DoptProvider } from '@dopt/react';
import { ReactElement, ReactNode, Suspense } from 'react';
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
        <DoptProvider
          userId="users-a611e66369fde320b693ad95843801e728b799e4a895bb54a810fc8d8317a502_MjE2Mw=="
          apiKey="blocks-28312c460558d766219ae7e8fd7a565d46efc6a6e1d316e48c9e15d30384ab60_MjE2MQ=="
          flowVersions={{ onboarding: 1 }}
        >
          <div className="relative flex min-h-screen flex-col">
            <SiteHeader />
            <div className="flex-1 container" id="top-level-layout">
              {children}
            </div>
            <SiteFooter />
            <Toaster />
          </div>
        </DoptProvider>
      </AccountProvider>
    </SessionProvider>
  );
}
