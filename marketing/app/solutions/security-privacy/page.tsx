import CTA from '@/components/cta/CTA';
import { DotBackground } from '@/components/landing-page/DotBackground';
import Intergrations from '@/components/landing-page/Integrations';
import Platform from '@/components/landing-page/old/Platform';
import { Metadata } from 'next';
import { ReactElement } from 'react';
import ComplySecurityPrivacy from './ComplySecurityPrivacy';
import TransformerSection from './TransformerSection';
import Hero from './hero';

export const metadata: Metadata = {
  metadataBase: new URL('https://assets.nucleuscloud.com/'),
  title: 'Easily comply with Data Privacy, Security and Compliance | Neosync',
  openGraph: {
    title: 'Easily comply with Data Privacy, Security and Compliance | Neosync',
    description:
      'Easily comply with laws like HIPAA, GDPR, and DPDP with de-identified and synthetic data that structurally and statistically looks just like your production data. ',
    url: 'https://neosync.dev/solutions/security-privacy',
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
          <ComplySecurityPrivacy />
        </div>
        <div className="bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto pt-20 lg:pt-40">
          <TransformerSection />
        </div>
        <div className=" bg-[#1E1E1E] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto pt-20 rounded-3xl mt-20 lg:mt-40 py-10 justify-center flex flex-col">
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
