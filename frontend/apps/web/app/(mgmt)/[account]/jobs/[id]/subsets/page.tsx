'use client';
import { PageProps } from '@/components/types';
import { ReactElement, use } from 'react';
import SubsetCard from './components/SubsetCard';

export default function Page(props: PageProps): ReactElement {
  const params = use(props.params);
  const id = params?.id ?? '';
  return (
    <div className="job-details-container">
      <SubsetCard jobId={id} />
    </div>
  );
}
