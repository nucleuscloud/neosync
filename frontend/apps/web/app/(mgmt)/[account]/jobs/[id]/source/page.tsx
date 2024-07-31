'use client';
import SubPageHeader from '@/components/headers/SubPageHeader';
import { PageProps } from '@/components/types';
import { ReactElement } from 'react';
import SourceConnectionCard from './components/SourceConnectionCard';
import TestEditor from './editortest';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  return (
    <div className="job-details-container">
      <SubPageHeader
        header="Source Connection"
        description="Manage a job's source connection. Click update at the bottom to persist any changes."
      />

      <div className="space-y-10">
        <TestEditor />
        <SourceConnectionCard jobId={id} />
      </div>
    </div>
  );
}
