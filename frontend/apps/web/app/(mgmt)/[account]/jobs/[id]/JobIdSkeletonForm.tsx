import { Skeleton } from '@/components/ui/skeleton';
import { ReactElement } from 'react';

export default function JobIdSkeletonForm(): ReactElement<any> {
  return (
    <div>
      <div className="space-y-4 pt-10">
        <div className="flex flex-row items-center justify-between">
          <Skeleton className="h-10 w-[200px]" />
          <div className="flex flex-row items-center gap-4">
            <Skeleton className="h-10 w-[200px]" />
            <Skeleton className="h-10 w-[200px]" />
            <Skeleton className="h-10 w-[200px]" />
            <Skeleton className="h-10 w-[200px]" />
          </div>
        </div>
        <Skeleton className="h-10 w-full" />
        <div className="flex flex-row items-center gap-4">
          <Skeleton className="h-10 w-[600px]" />
        </div>
        <div className="flex flex-row items-center gap-4 pt-6">
          <Skeleton className="h-40 w-3/4" />
          <Skeleton className="h-40 w-1/4" />
        </div>
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
      </div>
    </div>
  );
}
