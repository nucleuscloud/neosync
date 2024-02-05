import CTA from '@/components/cta/CTA';
import Hero from '@/components/landing-page/Hero';
import Intergrations from '@/components/landing-page/Integrations';
import Platform from '@/components/landing-page/Platform';
import UseNeosync from '@/components/landing-page/UseNeosync';
import ValueProps from '@/components/landing-page/Valueprops';
import { ReactElement } from 'react';

export default function Home(): ReactElement {
  return (
    <div>
      <div className="py-20 bg-[#FFFFFF] border-b border-b-gray-200">
        <Hero />
      </div>
      <div className=" bg-[#F5F5F5] lg:p-20 px-4">
        <div className="bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto pt-20 lg:pt-40">
          <ValueProps />
        </div>
        <div className="bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto pt-20 lg:py-40">
          <UseNeosync />
        </div>
        <div className=" bg-[#1E1E1E] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto rounded-3xl py-10">
          <div className="pt-4 lg:pt-20">
            <Platform />
          </div>
          <div className="pt-20 lg:pt-40">
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
