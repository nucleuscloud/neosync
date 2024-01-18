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
      <Hero />
      <ValueProps />
      <UseNeosync />
      <Transformers />
      <EngineeringTeams />
      <CTA />
    </div>
  );
}
