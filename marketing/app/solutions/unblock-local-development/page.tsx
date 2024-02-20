import CTA from '@/components/cta/CTA';
import APISection from '@/components/landing-page/APISection';
import { DotBackground } from '@/components/landing-page/DotBackground';
import Intergrations from '@/components/landing-page/Integrations';
import Platform from '@/components/landing-page/Platform';
import { Metadata } from 'next';
import { ReactElement } from 'react';
import LocalDevelopmentValueProps from './LocalDevelopmentValueProps';
import Hero from './hero';

export const metadata: Metadata = {
  metadataBase: new URL('https://assets.nucleuscloud.com/'),
  title: 'Neosync | Unblock Local Development',
  openGraph: {
    title: 'Neosync | Unblock Local Development',
    description:
      'Unblock local development by using Neosync to anonymize sensitive data and generate synthetic data so that developers can self-serve data locally.  ',
    url: 'https://neosync.dev/solutions/unblock-local-development',
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
      <div className="bg-[#FFFFFF] border-b border-b-gray-200">
        <DotBackground>
          <Hero />
        </DotBackground>
      </div>
      <div className=" bg-[#F5F5F5] px-4">
        <div className="bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto pt-20 lg:pt-40">
          <LocalDevelopmentValueProps />
        </div>
        <div className="bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto pt-20 lg:pt-40">
          <APISection />
        </div>

        <div className=" bg-[#1E1E1E] px-5 sm:px-10 md:px-20 lg:px-20 sm:mx-10 md:mx-20 lg:mx-40 max-w-[1800px] mx-auto rounded-3xl mt-20 lg:mt-40 py-10">
          <div className="pt-4 lg:pt-20">
            <Platform />
          </div>
          <div className="pt-20 lg:pt-20">
            <Intergrations />
          </div>
        </div>
        <div className="bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto py-10 lg:py-20">
          <CTA />
        </div>
      </div>
    </div>
  );
}
