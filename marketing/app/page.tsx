'use client';
import FeaturesGrid from '@/components/landing-page/FeaturesGrid';
import GitOpsSection from '@/components/landing-page/GitOps';
import Hero from '@/components/landing-page/Hero';
import Subset from '@/components/landing-page/Subset';
import TableTypes from '@/components/landing-page/TableTypes';
import Transformers from '@/components/landing-page/Transformers';
import UseCases from '@/components/landing-page/Usecases';
import { ReactElement } from 'react';

export default function Home(): ReactElement {
  return (
    <div>
      <Hero />
      <TableTypes />
      <FeaturesGrid />
      <UseCases />
      <Transformers />
      <Subset />
      <GitOpsSection />
    </div>
  );
}
