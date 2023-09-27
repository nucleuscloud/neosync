'use client';
import { PageProps } from '@/components/types';
import { useGetJob } from '@/libs/hooks/useGetJob';

import SubPageHeader from '@/components/headers/SubPageHeader';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { ReactElement } from 'react';
import JobScheduleCard from './components/ScheduleCard';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data, isLoading } = useGetJob(id);

  if (isLoading) {
    return (
      <div className="mt-10">
        <SkeletonForm />
      </div>
    );
  }

  return (
    <div className="job-details-container">
      <SubPageHeader
        header={data?.job?.name || ''}
        description={data?.job?.id || ''}
      />
      <div className="space-y-10">
        <JobScheduleCard jobId={id} />
      </div>
    </div>
  );
}
