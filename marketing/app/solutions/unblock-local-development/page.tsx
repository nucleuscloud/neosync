import CTA from '@/components/cta/CTA';
import APISection from '@/components/landing-page/APISection';
import Intergrations from '@/components/landing-page/Integrations';
import Platform from '@/components/landing-page/Platform';
import { Metadata } from 'next';
import { ReactElement } from 'react';
import LocalDevelopmentValueProps from './LocalDevelopmentValueProps';
import Hero from './hero';

export const metadata: Metadata = {
  metadataBase: new URL('https://assets.nucleuscloud.com/'),
  title: 'Unblock Local Development | Neosync',
  openGraph: {
    title: 'Unblock Local Development | Neosync',
    description:
      'Unblock local development by using Neosync to anonymize sensitive data and generate synthetic data so that developers can self-serve data locally.  ',
    url: 'https://www.neosync.dev/solutions/unblock-local-development',
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
    canonical: 'https://www.neosync.dev/solutions/unblock-local-development',
  },
};

export default function Page(): ReactElement {
  return (
    <div>
      <div className="pb-8 lg:pb-20 pt-20">
        <Hero />
      </div>
      <div
        className="bg-gradient-to-tr from-[#f8f8f8] to-[#eaeaea] mt-20 lg:mt-20 pt-20 p-10 rounded-xl border border-gray-300"
        id="value-props-section"
      >
        <LocalDevelopmentValueProps />
      </div>
      <div className=" pt-20 lg:pt-40">
        <APISection />
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
