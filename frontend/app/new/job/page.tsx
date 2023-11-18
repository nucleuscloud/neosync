'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
import { nanoid } from 'nanoid';
import { useRouter } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';

export default function NewJob({ params }: PageProps): ReactElement {
  const [sessionToken] = useState(params?.sessionToken ?? nanoid());
  const router = useRouter();
  useEffect(() => {
    router.push(`/new/job/define?sessionId=${sessionToken}`);
  }, [sessionToken]);

  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Create a new Job"
          description="Define a new job to move, transform, or scan data"
        />
      }
    >
      <div></div>
    </OverviewContainer>
  );
}
