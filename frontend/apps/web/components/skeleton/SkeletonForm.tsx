import { ReactElement } from 'react';
import { Skeleton } from '../ui/skeleton';

interface Props {}

export default function SkeletonForm(props: Props): ReactElement<any> {
  const {} = props;
  return (
    <div className="px-32 space-y-10">
      <div className="flex flex-row items-center justify-between">
        <Skeleton className="h-10 w-[400px]" />
        <div className="flex flex-row items-center gap-4">
          <Skeleton className="h-10 w-[200px]" />
          <Skeleton className="h-10 w-[200px]" />
        </div>
      </div>
      <div className="space-y-10">
        <Skeleton className="h-12 w-full" />
        <Skeleton className="h-12 w-full" />
        <Skeleton className="h-12 w-full" />
        <Skeleton className="h-12 w-full" />
        <Skeleton className="h-12 w-full" />
      </div>
    </div>
  );
}
