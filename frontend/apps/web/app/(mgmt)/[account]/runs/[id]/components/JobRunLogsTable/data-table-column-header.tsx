import { Column } from '@tanstack/react-table';

import { cn } from '@/libs/utils';

interface DataTableColumnHeaderProps<TData, TValue>
  extends React.HTMLAttributes<HTMLDivElement> {
  column: Column<TData, TValue>;
  title: string;
}

export function DataTableColumnHeader<TData, TValue>({
  title,
  className,
}: DataTableColumnHeaderProps<TData, TValue>) {
  return (
    <div className={cn(className, 'text-xs')}>
      <p>{title}</p>
    </div>
  );
}
