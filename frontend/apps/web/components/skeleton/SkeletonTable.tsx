import { ReactElement } from 'react';
import { Skeleton } from '../ui/skeleton';

interface Props {}

export default function SkeletonTable(props: Props): ReactElement {
  const {} = props;
  return (
    <div className="space-y-4">
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
    </div>
  );
}
