'use client';
import { default as Features } from '@/components/landing-page/Features';
import GitOpsSection from '@/components/landing-page/GitOps';
import Hero from '@/components/landing-page/Hero';
import Subset from '@/components/landing-page/Subset';
import TableTypes from '@/components/landing-page/TableTypes';
import Transformers from '@/components/landing-page/Transformers';
import UseNeosync from '@/components/landing-page/UseNeosync';
import UseCases from '@/components/landing-page/Usecases';
import ValueProps from '@/components/landing-page/Valueprops';
import { ReactElement } from 'react';

export default function Home(): ReactElement {
  return (
    <div>
      <Hero />
      <ValueProps />
      <UseNeosync />
      <Features />
      <TableTypes />
      <UseCases />
      <Transformers />
      <Subset />
      <GitOpsSection />
    </div>
  );
}
