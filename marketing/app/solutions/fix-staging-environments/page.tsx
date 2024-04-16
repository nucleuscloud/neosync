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
  alternates: {
    canonical: 'https://www.neosync.dev/solutions/fix-staging-environments',
  },
};

export default function Page(): ReactElement {
  return (
    <div>
      <Hero />
      <div
        className="bg-gradient-to-tr from-[#f8f8f8] to-[#eaeaea] mt-20 lg:mt-20 pt-20 p-10 rounded-xl border border-gray-300"
        id="value-props-section"
      >
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
