'use client';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
import { ReactElement } from 'react';
import DestinationConnectionCard from './components/DestinationConnectionCard';
import SourceConnectionCard from './components/SourceConnectionCard';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';

  return (
    <div className="job-details-container">
      <PageHeader header="Connections" description="Manage job connections" />

      <div className="space-y-10">
        <SourceConnectionCard jobId={id} />
        <DestinationConnectionCard jobId={id} />
      </div>
    </div>
  );
}
