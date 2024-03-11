import CTA from '@/components/cta/CTA';
import Intergrations from '@/components/landing-page/Integrations';
import Platform from '@/components/landing-page/Platform';
import { Metadata } from 'next';
import { ReactElement } from 'react';
import FixBrokenStaging from './FixBrokenStaging';
import Hero from './hero';

export const metadata: Metadata = {
  metadataBase: new URL('https://assets.nucleuscloud.com/'),
  title: 'Fix Staging Environments | Neosync',
  openGraph: {
    title: 'Fix Staging Environments | Neosync',
    description:
      'Fix broken staging environments and catch bugs before production. ',
    url: 'https://neosync.dev/solutions/fix-staging-environments',
    siteName: 'Neosync',
    images: [
      {
        url: '/neosync/marketingsite/mainOGHero.svg',
        width: 1200,
        height: 630,
        alt: 'mainOG',
      },
    ],
    locale: 'en_US',
    type: 'website',
  },
};

export default function Page(): ReactElement {
  return (
    <div>
      <Hero />
      <div className="pt-20 lg:pt-40">
        <FixBrokenStaging />
      </div>
      <div className="pt-20 lg:pt-40 flex justify-center" id="platform-section">
        <Platform />
      </div>
      <div className="pt-20 lg:pt-20">
        <Intergrations />
      </div>
      <div className=" py-10 lg:py-20">
        <CTA />
      </div>
    </div>
  );
}
