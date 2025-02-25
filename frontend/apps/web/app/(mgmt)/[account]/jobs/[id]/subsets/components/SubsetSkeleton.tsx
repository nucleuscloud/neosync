import { Skeleton } from '@/components/ui/skeleton';
import { ReactElement } from 'react';

export default function SubsetSkeleton(): ReactElement {
  return (
    <div className="space-y-4">
      <Skeleton className="h-20 w-3/4" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
      <div className="flex flex-row justify-between items-center">
        <Skeleton className="h-10 w-full" />
        <div className="flex flex-row items-center gap-4">
          <Skeleton className="h-10 w-[100px]" />
          <Skeleton className="h-10 w-[100px]" />
          <Skeleton className="h-10 w-[100px]" />
        </div>
      </div>
      <Skeleton className="h-20 w-full" />
      <Skeleton className="h-20 w-full" />
    </div>
  );
}
