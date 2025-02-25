'use client';
import {
  StickyHeaderTable,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { cn } from '@/libs/utils';
import { Cross2Icon } from '@radix-ui/react-icons';
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFacetedMinMaxValues,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import { useVirtualizer } from '@tanstack/react-virtual';
import React, { ReactElement } from 'react';
import ButtonText from '../ButtonText';
import { Button } from '../ui/button';

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  ConnectionAlert: ReactElement<any>;
  TestConnectionButton?: ReactElement<any>;
}

export default function PermissionsDataTable<TData, TValue>({
  columns,
  data,
  ConnectionAlert,
  TestConnectionButton,
}: DataTableProps<TData, TValue>): ReactElement<any> {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getFacetedMinMaxValues: getFacetedMinMaxValues(),
    initialState: {
      sorting: [{ id: 'schemaTable', desc: true }],
      columnVisibility: {
        schema: false,
        table: false,
      },
    },
  });

  const { rows } = table.getRowModel();

  const tableContainerRef = React.useRef<HTMLDivElement>(null);

  const rowVirtualizer = useVirtualizer({
    count: rows.length,
    estimateSize: () => 33,
    getScrollElement: () => tableContainerRef.current,
    overscan: 5,
  });

  return (
    (<div className="space-y-4">
      <div className="flex flex-col lg:flex-row items-center gap-2 lg:gap-4 justify-between">
        <div>{ConnectionAlert}</div>
        <div className="flex flex-col lg:flex-row items-center gap-2">
          {TestConnectionButton && TestConnectionButton}
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
        className={cn(
          'rounded-md border min-h-[145px] max-h-[500px] relative border-gray-300 dark:border-gray-700 overflow-hidden',
          rows.length > 0 && 'overflow-auto'
        )}
        ref={tableContainerRef}
      >
        <StickyHeaderTable>
          <TableHeader className="bg-gray-100 dark:bg-gray-800 sticky top-0 z-10 grid">
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow
                key={headerGroup.id}
                className="flex justify-between w-full px-2"
              >
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      key={header.id}
                      style={{ width: header.getSize() }}
                      colSpan={header.colSpan}
                      className="flex w-full"
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
            className="grid relative"
            style={{
              height: `${rowVirtualizer.getTotalSize()}px`, //tells scrollbar how big the table is
            }}
          >
            {rows.length === 0 && (
              <TableRow className="flex justify-center items-center py-10 text-gray-500 text-sm">
                <td>No permissions found for the connection or filter(s).</td>
              </TableRow>
            )}
            {rowVirtualizer.getVirtualItems().map((virtualRow) => {
              const row = rows[virtualRow.index];
              return (
                (<TableRow
                  data-index={virtualRow.index} //needed for dynamic row height measurement
                  ref={node => {
                    rowVirtualizer.measureElement(node);
                  }} //measure dynamic row height
                  key={row.id}
                  style={{
                    transform: `translateY(${virtualRow.start}px)`,
                  }}
                  className="flex absolute justify-between w-full px-2"
                >
                  {row.getVisibleCells().map((cell) => {
                    return (
                      <td
                        key={cell.id}
                        className="py-2 text-sm flex"
                        style={{
                          width: cell.column.getSize(),
                        }}
                      >
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </td>
                    );
                  })}
                </TableRow>)
              );
            })}
          </TableBody>
        </StickyHeaderTable>
      </div>
    </div>)
  );
}
