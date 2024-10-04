'use client';

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Row } from '@tanstack/react-table';

import { DotsHorizontalIcon } from '@radix-ui/react-icons';

interface DataTableRowActionsProps<TData> {
  row: Row<TData>;
  onViewSelectClicked(): void;
}

export function DataTableRowActions<TData>({
  onViewSelectClicked,
}: DataTableRowActionsProps<TData>) {
  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger className="hover:bg-gray-100 dark:hover:bg-gray-800 py-1 px-2 rounded-lg">
        <DotsHorizontalIcon className="h-4 w-4" />
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuItem
          className="cursor-pointer"
          onClick={() => onViewSelectClicked()}
        >
          Show Select Query
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
