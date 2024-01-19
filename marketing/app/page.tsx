'use client';
import CTA from '@/components/cta/CTA';
import EngineeringTeams from '@/components/landing-page/EngineeringTeams';
import Hero from '@/components/landing-page/Hero';
import Transformers from '@/components/landing-page/Transformers';
import UseNeosync from '@/components/landing-page/UseNeosync';
import ValueProps from '@/components/landing-page/Valueprops';
import { ReactElement } from 'react';

export default function Home(): ReactElement {
  return (
    <div>
      <div className="py-20 bg-[#FFFFFF] border-b border-b-gray-200">
        <Hero />
      </div>
      <div className="bg-[#F5F5F5] px-5 sm:px-10 md:px-20 lg:px-40 max-w-[1800px] mx-auto">
        <div className="pt-20 lg:pt-40">
          <ValueProps />
        </div>
        <div className="pt-20 lg:pt-40">
          <UseNeosync />
        </div>
        <div className="pt-20 lg:pt-40">
          <Transformers />
        </div>
        <div className="pt-20 lg:pt-40">
          <EngineeringTeams />
        </div>
        <div className="py-20 lg:pt-40">
          <CTA />
        </div>
      </div>
    </div>
  );
}
