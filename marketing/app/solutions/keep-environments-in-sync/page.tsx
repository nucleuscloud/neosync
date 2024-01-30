import CTA from '@/components/cta/CTA';
import Intergrations from '@/components/landing-page/Integrations';
import Platform from '@/components/landing-page/Platform';
import { Metadata } from 'next';
import { ReactElement } from 'react';
import JobSchedules from './JobShedules';
import KeepEnvironmentsInSync from './KeepEnvironmentsInSync';
import Hero from './hero';

export const metadata: Metadata = {
  metadataBase: new URL('https://assets.nucleuscloud.com/'),
  openGraph: {
    title: 'Neosync | Keep Environments In Sync',
    description:
      'Keep environments in sync by anonymizing sensitive data and generating synthetic data and syncing it across environemnts. ',
    url: 'https://neosync.dev/keep-environments-in-sync',
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
      <div>
        <div className="py-20 bg-[#FFFFFF] border-b border-b-gray-200 mx-6 lg:mx-40">
          <Hero />
        </div>
        <div className="bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto pt-20 lg:pt-40">
          <KeepEnvironmentsInSync />
        </div>
        <div className="bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto pt-20 lg:pt-40">
          <JobSchedules />
        </div>
        <div className=" bg-[#F5F5F5] lg:p-20 px-4">
          <div className=" bg-[#1E1E1E] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto rounded-3xl py-10">
            <div className="pt-4 lg:pt-20">
              <Platform />
            </div>
            <div className="pt-20 lg:pt-40">
              <Intergrations />
            </div>
          </div>
        </div>
        <div className="bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto py-10 lg:py-20">
          <CTA />
        </div>
      </div>
    </div>
  );
}
