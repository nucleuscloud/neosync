import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { cn } from '@/libs/utils';
import {
  ArrowDownIcon,
  ArrowUpIcon,
  MagnifyingGlassIcon,
} from '@radix-ui/react-icons';
import { Column } from '@tanstack/react-table';

interface DataTableColumnHeaderProps<TData, TValue>
  extends React.HTMLAttributes<HTMLDivElement> {
  column: Column<TData, TValue>;
  title: string;
}

export function DataTableColumnHeader<TData, TValue>({
  column,
  title,
  className,
}: DataTableColumnHeaderProps<TData, TValue>) {
  return (
    <div className={cn('flex items-center', className)}>
      <div className="flex flex-row space-x-2 items-center">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="-ml-3 data-[state=open]:bg-accent hover:border hover:border-gray-400 text-nowrap"
            >
              <span>{title}</span>
              {column.getIsSorted() === 'desc' ? (
                <ArrowDownIcon className="ml-2 h-4 w-4" />
              ) : column.getIsSorted() === 'asc' ? (
                <ArrowUpIcon className="ml-2 h-4 w-4" />
              ) : (
                <MagnifyingGlassIcon className="ml-2 h-4 w-4" />
              )}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start">
            <Input
              type="text"
              value={(column.getFilterValue() ?? '') as string}
              onChange={(e) => column.setFilterValue(e.target.value)}
              placeholder={`Search...`}
              className="w-36 border rounded"
            />
            {column.getCanSort() && (
              <>
                <DropdownMenuItem onClick={() => column.toggleSorting(false)}>
                  <ArrowUpIcon className="mr-2 h-3.5 w-3.5 text-muted-foreground/70" />
                  Asc
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => column.toggleSorting(true)}>
                  <ArrowDownIcon className="mr-2 h-3.5 w-3.5 text-muted-foreground/70" />
                  Desc
                </DropdownMenuItem>
              </>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  );
}
