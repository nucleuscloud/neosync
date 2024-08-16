import SiteFooter from '@/components/SiteFooter';
import WelcomeDialog from '@/components/onboarding-checklist/WelcomeDialog';
import AccountProvider from '@/components/providers/account-provider';
import ConnectProvider from '@/components/providers/connect-provider';
import { KoalaIdentifier } from '@/components/providers/koala-provider';
import { PostHogIdentifier } from '@/components/providers/posthog-provider';
import TanstackQueryProvider from '@/components/providers/query-provider';
import { SessionProvider } from '@/components/providers/session-provider';
import SiteHeader from '@/components/site-header/SiteHeader';
import { Toaster } from '@/components/ui/sonner';
import { ReactElement, ReactNode, Suspense } from 'react';
import { auth } from './api/auth/[...nextauth]/auth';
import { getSystemAppConfig } from './api/config/config';

interface Props {
  children: ReactNode;
}
export default async function BaseLayout(props: Props): Promise<ReactElement> {
  const { children } = props;
  const session = await auth();
  const { publicNeosyncApiBaseUrl } = getSystemAppConfig();

  return (
    <ConnectProvider apiBaseUrl={publicNeosyncApiBaseUrl}>
      <TanstackQueryProvider>
        <SessionProvider session={session}>
          <AccountProvider>
            <Suspense>
              <PostHogIdentifier />
            </Suspense>
            <Suspense>
              <KoalaIdentifier />
            </Suspense>
            <div className="relative flex min-h-screen flex-col">
              <SiteHeader />
              <div className="flex-1 container" id="top-level-layout">
                {children}
              </div>
              <SiteFooter />
              {/* https://sonner.emilkowal.ski/styling for styling documentation */}
              <Toaster richColors closeButton />
              {/* <OnboardingChecklist /> */}
              <WelcomeDialog />
            </div>
          </AccountProvider>
        </SessionProvider>
      </TanstackQueryProvider>
    </ConnectProvider>
  );
}
