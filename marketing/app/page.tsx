import CTA from '@/components/cta/CTA';
import DeploymentOptions from '@/components/landing-page/DeploymentOptions';
import GitOpsSection from '@/components/landing-page/GitOps';
import Hero from '@/components/landing-page/Hero';
import Intergrations from '@/components/landing-page/Integrations';
import Platform from '@/components/landing-page/Platform';
import UseHostedNeosync from '@/components/landing-page/UseHostedNeosync';
import ValueProps from '@/components/landing-page/Valueprops';
import { Metadata } from 'next';
import { ReactElement } from 'react';

export const metadata: Metadata = {
  metadataBase: new URL('https://assets.nucleuscloud.com/'),
  title: 'Neosync | Synthetic Data Orchestration',
  openGraph: {
    title: 'Neosync',
    description: 'Open Source Synthetic Data Orchestration',
    url: 'https://www.neosync.dev',
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
    canonical: 'https://www.neosync.dev',
  },
};

export default function Home(): ReactElement {
  return (
    <div>
      <Hero />
      <div
        className="bg-gradient-to-tr from-[#f8f8f8] to-[#eaeaea] mt-20 pt-20 p-10 rounded-xl border border-gray-300"
        id="value-props-section"
      >
        <ValueProps />
      </div>
      <div className="pt-20 lg:pt-40 flex justify-center" id="platform-section">
        <Platform />
      </div>
      <div className="pt-20">
        <Intergrations />
      </div>
      <div className="py-10 lg:pt-40">
        <DeploymentOptions />
      </div>
      <div className=" pt-20 lg:pt-40">
        <UseHostedNeosync />
      </div>
      <div className="py-10 lg:py-20">
        <GitOpsSection />
      </div>
      <div className="px-5 lg-px-2 py-10 lg:py-20">
        <CTA />
      </div>
    </div>
  );
}
