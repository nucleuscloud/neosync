'use client';
import SubPageHeader from '@/components/headers/SubPageHeader';
import { PageProps } from '@/components/types';
import { ReactElement, use } from 'react';
import SourceConnectionCard from './components/SourceConnectionCard';

export default function Page(props: PageProps): ReactElement<any> {
  const params = use(props.params);
  const id = params?.id ?? '';
  return (
    <div className="job-details-container flex flex-col gap-5">
      <SubPageHeader
        header="Source Connection"
        description="Manage a job's source connection. Click update at the bottom to persist any changes."
      />

      <div className="space-y-10">
        <SourceConnectionCard jobId={id} />
      </div>
    </div>
  );
}
