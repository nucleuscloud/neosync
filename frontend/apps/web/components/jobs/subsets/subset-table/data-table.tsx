'use client';

import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';

import ButtonText from '@/components/ButtonText';
import { Button } from '@/components/ui/button';
import {
  StickyHeaderTable,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Cross2Icon } from '@radix-ui/react-icons';
import { useVirtualizer } from '@tanstack/react-virtual';
import { useRef } from 'react';
interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
}

export function DataTable<TData, TValue>({
  columns,
  data,
}: DataTableProps<TData, TValue>) {
  const table = useReactTable({
    data,
    columns,
    enableRowSelection: true,
    initialState: {
      columnVisibility: {
        schema: false,
        table: false,
      },
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
  });

  const { rows } = table.getRowModel();

  const tableContainerRef = useRef<HTMLDivElement>(null);

  const rowVirtualizer = useVirtualizer({
    count: rows.length,
    estimateSize: () => 33,
    getScrollElement: () => tableContainerRef.current,
    measureElement:
      typeof window !== 'undefined' &&
      navigator.userAgent.indexOf('Firefox') === -1
        ? (element) => element?.getBoundingClientRect().height
        : undefined,
    overscan: 5,
  });

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-row items-center gap-2 justify-end">
        <div className="flex flex-row items-center gap-2">
          <Button
            disabled={table.getState().columnFilters.length === 0}
            type="button"
            variant="outline"
            className="px-2 lg:px-3"
            onClick={() => table.resetColumnFilters()}
          >
            <ButtonText
              leftIcon={<Cross2Icon className="h-3 w-3" />}
              text="Clear filters"
            />
          </Button>
        </div>
      </div>
      <div
        className="rounded-md border max-h-[500px] relative overflow-x-auto"
        ref={tableContainerRef}
      >
        <StickyHeaderTable>
          <TableHeader className="bg-gray-100 dark:bg-gray-800 sticky top-0 z-10 flex w-full px-2">
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow
                key={headerGroup.id}
                className="flex flex-row items-center justify-between w-full"
                id="table-header-row"
              >
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      key={header.id}
                      style={{ minWidth: `${header.column.getSize()}px` }}
                      colSpan={header.colSpan}
                      className="flex items-center"
                    >
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                    </TableHead>
                  );
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody
            style={{
              height: `${rowVirtualizer.getTotalSize()}px`, //tells scrollbar how big the table is
            }}
          >
            {rows.length === 0 && (
              <TableRow className="flex justify-center items-center py-10 text-gray-500">
                <td>No schema(s) or table(s) found.</td>
              </TableRow>
            )}
            {rowVirtualizer.getVirtualItems().map((virtualRow) => {
              const row = rows[virtualRow.index];
              return (
                <TableRow
                  data-index={virtualRow.index} //needed for dynamic row height measurement
                  ref={(node) => rowVirtualizer.measureElement(node)} //measure dynamic row height
                  key={row.id}
                  style={{
                    transform: `translateY(${virtualRow.start}px)`,
                  }}
                  className="items-center flex absolute w-full justify-between px-2"
                >
                  {row.getVisibleCells().map((cell) => {
                    return (
                      <td
                        key={cell.id}
                        className="py-2"
                        style={{
                          minWidth: cell.column.getSize(),
                        }}
                      >
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </td>
                    );
                  })}
                </TableRow>
              );
            })}
          </TableBody>
        </StickyHeaderTable>
      </div>
    </div>
  );
}
