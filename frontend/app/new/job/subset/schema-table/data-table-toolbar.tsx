'use client';

import { Table } from '@tanstack/react-table';

import { Button } from '@/components/ui/button';
import { UpdateIcon } from '@radix-ui/react-icons';

interface DataTableToolbarProps<TData> {
  table: Table<TData>;
  onClearFilters: () => void;
}

export function DataTableToolbar<TData>({
  table,
  onClearFilters,
}: DataTableToolbarProps<TData>) {
  return (
    <div className="flex items-center justify-between">
      <Button
        variant="outline"
        type="button"
        onClick={() => {
          table.setColumnFilters([]);
          onClearFilters();
        }}
      >
        Clear filters
        <UpdateIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
      </Button>
    </div>
  );
}
