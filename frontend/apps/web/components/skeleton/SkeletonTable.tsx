import { ReactElement } from 'react';
import { Skeleton } from '../ui/skeleton';

interface Props {}

export default function SkeletonTable(props: Props): ReactElement {
  const {} = props;
  return (
    <div className="space-y-4">
      <div className="flex flex-row items-center justify-between">
        <Skeleton className="h-10 w-[200px]" />
        <Skeleton className="h-10 w-[100px]" />
      </div>
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
      <div className="flex justify-end">
        <Skeleton className="h-10 w-[400px]" />
      </div>
    </div>
  );
}
