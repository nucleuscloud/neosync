'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import { PageProps } from '@/components/types';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetJob } from '@/libs/hooks/useGetJob';
import { ReactElement } from 'react';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data, isLoading } = useGetJob(id);
  // const { toast } = useToast();
  if (!id) {
    return <div>Not Found</div>;
  }
  if (isLoading) {
    return <Skeleton className="w-[100px] h-[20px] rounded-full" />;
  }

  return (
    <OverviewContainer Header={<div>Header todo</div>}>
      <div className="job-details-container">
        <div>
          <div>
            <div>{data?.job?.id}</div>
          </div>
        </div>
      </div>
    </OverviewContainer>
  );
}
