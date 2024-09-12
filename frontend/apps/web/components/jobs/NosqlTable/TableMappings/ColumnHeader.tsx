import {
  ArrowDownIcon,
  ArrowUpIcon,
  CaretSortIcon,
} from '@radix-ui/react-icons';
import { Column } from '@tanstack/react-table';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { cn } from '@/libs/utils';
import { FaSearch } from 'react-icons/fa';

interface ColumnHeaderProps<TData, TValue>
  extends React.HTMLAttributes<HTMLDivElement> {
  column: Column<TData, TValue>;
  title: string;
}

export function ColumnHeader<TData, TValue>({
  column,
  title,
  className,
}: ColumnHeaderProps<TData, TValue>) {
  return (
    <div className="flex flex-row gap-2 items-center justify-start">
      {column.getCanFilter() && (
        <div>
          <div className="relative">
            <Input
              type="text"
              value={(column.getFilterValue() ?? '') as string}
              onChange={(e) => column.setFilterValue(e.target.value)}
              placeholder={title}
              className="border border-gray-300 dark:border-gray-700 bg-white dark:bg-transparent text-xs h-8 pl-8"
            />
            <FaSearch className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
          </div>
        </div>
      )}
      {!column.getCanFilter() && (
        <span className={cn(className, 'text-xs')}>{title}</span>
      )}
      {column.getCanSort() && (
        <div>
          <Button
            type="button"
            onClick={() => {
              const sorted = column.getIsSorted();

              if (!sorted) {
                column.toggleSorting(false);
              } else if (sorted === 'asc') {
                column.toggleSorting(true);
              } else if (sorted === 'desc') {
                column.toggleSorting(undefined);
              }
            }}
            variant="ghost"
            className="px-1"
          >
            {column.getIsSorted() === 'desc' ? (
              <ArrowDownIcon className="h-4 w-4" />
            ) : column.getIsSorted() === 'asc' ? (
              <ArrowUpIcon className="h-4 w-4" />
            ) : (
              <CaretSortIcon className="h-4 w-4" />
            )}
          </Button>
        </div>
      )}
    </div>
  );
}
