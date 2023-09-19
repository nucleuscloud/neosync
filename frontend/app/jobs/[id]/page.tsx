'use client';
import { PageProps } from '@/components/types';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetJob } from '@/libs/hooks/useGetJob';

import SubPageHeader from '@/components/headers/SubPageHeader';
import { ReactElement } from 'react';
import JobScheduleCard from './components/ScheduleCard';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data, isLoading } = useGetJob(id);

  if (isLoading) {
    return <Skeleton />;
  }

  return (
    <div className="job-details-container">
      <SubPageHeader
        header={data?.job?.name || ''}
        description="View job details"
      />
      <div className="space-y-10">
        <JobScheduleCard jobId={id} />
      </div>
    </div>
  );
}
