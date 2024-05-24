import CTA from '@/components/cta/CTA';
import Intergrations from '@/components/landing-page/Integrations';
import Platform from '@/components/landing-page/Platform';
import { Metadata } from 'next';
import { ReactElement } from 'react';
import ComplySecurityPrivacy from './ComplySecurityPrivacy';
import TransformerSection from './TransformerSection';
import Hero from './hero';

export const metadata: Metadata = {
  metadataBase: new URL('https://assets.nucleuscloud.com/'),
  title: 'Easily comply with Data Privacy, Security and Compliance | Neosync',
  description:
    'Easily comply with laws like HIPAA, GDPR, and DPDP with de-identified and synthetic data that structurally and statistically looks just like your production data. ',
  openGraph: {
    title: 'Easily comply with Data Privacy, Security and Compliance | Neosync',
    description:
      'Easily comply with laws like HIPAA, GDPR, and DPDP with de-identified and synthetic data that structurally and statistically looks just like your production data. ',
    url: 'https://www.neosync.dev/solutions/security-privacy',
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
    canonical: 'https://www.neosync.dev/solutions/security-privacy',
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
        <ComplySecurityPrivacy />
      </div>
      <div className=" pt-20 lg:pt-40">
        <TransformerSection />
      </div>
      <div className="pt-20 flex justify-center" id="platform-section">
        <Platform />
      </div>
      <div className="pt-20 ">
        <Intergrations />
      </div>
      <div className="py-10 lg:py-20">
        <CTA />
      </div>
    </div>
  );
}
