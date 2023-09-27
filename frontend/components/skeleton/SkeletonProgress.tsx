import { ReactElement } from 'react';
import { Skeleton } from '../ui/skeleton';

interface Props {}

export default function SkeletonProgress(props: Props): ReactElement {
  const {} = props;
  return (
    <div className="space-y-4">
      <div className="flex flew-row space-x-4">
        <Skeleton className="w-10 h-10 rounded-full" />
        <Skeleton className="w-1/3 h-10" />
      </div>
      <div className="flex flew-row space-x-4">
        <Skeleton className="w-10 h-10 rounded-full" />
        <Skeleton className="w-1/3 h-10" />
      </div>
      <div className="flex flew-row space-x-4">
        <Skeleton className="w-10 h-10 rounded-full" />
        <Skeleton className="w-1/3 h-10" />
      </div>
      <div className="flex flew-row space-x-4">
        <Skeleton className="w-10 h-10 rounded-full" />
        <Skeleton className="w-1/3 h-10" />
      </div>
    </div>
  );
}
