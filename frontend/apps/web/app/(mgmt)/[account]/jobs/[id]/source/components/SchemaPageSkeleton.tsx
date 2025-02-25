import { Skeleton } from '@/components/ui/skeleton';
import { ReactElement } from 'react';

export default function SchemaPageSkeleton(): ReactElement<any> {
  return (
    <div className="space-y-10">
      <Skeleton className="w-full h-10" />
      <div className="flex flex-row items-center gap-4">
        <Skeleton className="h-40 w-1/2" />
        <Skeleton className="h-40 w-1/2" />
      </div>
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
    </div>
  );
}
