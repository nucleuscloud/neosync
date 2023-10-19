'use client';
import { PageProps } from '@/components/types';
import { ReactElement } from 'react';
import SubsetCard from './components/SubsetCard';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  return (
    <div className="job-details-container">
      <div className="space-y-10">
        <SubsetCard jobId={id} />
      </div>
    </div>
  );
}
