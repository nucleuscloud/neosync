import { ReactElement } from 'react';
import { Skeleton } from '../ui/skeleton';

interface Props {}

export default function SkeletonForm(props: Props): ReactElement {
  const {} = props;
  return (
    <div className="space-y-10">
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-12 w-full" />
    </div>
  );
}
