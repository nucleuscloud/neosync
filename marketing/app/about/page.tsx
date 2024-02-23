import { Metadata } from 'next';
import { DotBackground } from './DotBackground';
import Hero from './Hero';
import InvestorSection from './Investors';
import OpenSource from './OpenSource';
import Values from './Values';

export const metadata: Metadata = {
  metadataBase: new URL('https://assets.nucleuscloud.com/'),
  title: 'Neosync | About',
  openGraph: {
    title: 'Neosync | About',
    description: `Learn more about Neosync&apos;s mission and values `,
    url: 'https://neosync.dev/solutions/keep-environments-in-sync',
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

export default function About() {
  return (
    <div className="bg-[#FFFFFF] border-b border-b-gray-200 pt-5">
      <DotBackground>
        <Hero />
      </DotBackground>
      <div className=" bg-[#F5F5F5] px-4">
        <div className="bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto pt-20">
          <Values />
        </div>
        {/* <div className=" bg-[#1E1E1E] px-5 sm:px-10 md:px-20 lg:px-20 sm:mx-10 md:mx-20 lg:mx-40 max-w-[1800px] mx-auto rounded-3xl mt-20 lg:mt-40 py-10">
          <TeamSection />
        </div> */}
        <div className=" bg-[#1E1E1E] px-5 sm:px-10 md:px-20 lg:px-20 sm:mx-10 md:mx-20 lg:mx-40 max-w-[1800px] mx-auto rounded-3xl mt-20 lg:mt-40 py-10">
          <OpenSource />
        </div>
        <div className="bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto pt-20 lg:pt-40">
          <InvestorSection />
        </div>
      </div>
    </div>
  );
}
