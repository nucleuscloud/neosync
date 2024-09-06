'use client';

import {
  ColumnDef,
  Table,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';

import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  StickyHeaderTable,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { cn } from '@/libs/utils';
import { LogLevel } from '@neosync/sdk';
import { ReloadIcon } from '@radix-ui/react-icons';
import { useVirtualizer } from '@tanstack/react-virtual';
import { useRef } from 'react';

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];

  getFuzzyFilterValue(table: Table<TData>): string;
  setFuzzyFilterValue(table: Table<TData>, value: string): void;

  selectedLogLevel: LogLevel;
  setSelectedLogLevel(newval: LogLevel): void;
  isLoading: boolean;
}

export function DataTable<TData, TValue>({
  columns,
  data,
  getFuzzyFilterValue,
  setFuzzyFilterValue,
  selectedLogLevel,
  setSelectedLogLevel,
  isLoading,
}: DataTableProps<TData, TValue>) {
  const table = useReactTable({
    data,
    columns,
    enableRowSelection: false,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
  });

  const { rows } = table.getRowModel();

  const tableContainerRef = useRef<HTMLDivElement>(null);

  const rowVirtualizer = useVirtualizer({
    count: rows.length,
    estimateSize: () => 33,
    getScrollElement: () => tableContainerRef.current,
    overscan: 5,
  });
  return (
    <div className="space-y-4">
      <div className="flex lg:w-1/2 gap-2 flex-col lg:flex-row">
        <Input
          placeholder="Search logs..."
          value={getFuzzyFilterValue(table)}
          onChange={(e) => setFuzzyFilterValue(table, e.target.value)}
        />
        <div className="flex flex-row gap-2 items-center">
          <div>
            <p className="font-light text-xs">Log Level</p>
          </div>
          <div className="flex w-full">
            <Select
              onValueChange={(value) =>
                setSelectedLogLevel(parseInt(value, 10))
              }
              value={selectedLogLevel.toString()}
            >
              <SelectTrigger>
                <SelectValue className="w-[500px]" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={LogLevel.UNSPECIFIED.toString()}>
                  {'Any'}
                </SelectItem>
                <SelectItem value={LogLevel.INFO.toString()}>
                  {LogLevel[LogLevel.INFO]}
                </SelectItem>
                <SelectItem value={LogLevel.WARN.toString()}>
                  {LogLevel[LogLevel.WARN]}
                </SelectItem>
                <SelectItem value={LogLevel.ERROR.toString()}>
                  {LogLevel[LogLevel.ERROR]}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
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
          <TableHeader className="bg-gray-100 dark:bg-gray-800 sticky top-0 z-10 px-2 grid">
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow
                key={headerGroup.id}
                className="flex flex-row w-full px-2"
              >
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      key={header.id}
                      style={{ minWidth: `${header.column.getSize()}px` }}
                      colSpan={header.colSpan}
                      className="flex items-center px-2"
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
            className="grid"
            style={{
              height: `${rowVirtualizer.getTotalSize()}px`, //tells scrollbar how big the table is
            }}
          >
            {rows.length === 0 && (
              <TableRow className="flex justify-center items-center py-10 text-gray-500">
                <td>
                  <div className="flex w-full flex-row gap-2 items-center">
                    {isLoading ? (
                      <ReloadIcon className="h-4 w-4 animate-spin" />
                    ) : null}
                    <p>
                      {isLoading
                        ? 'Waiting for logs to load...'
                        : 'No logs found for the given query'}
                    </p>
                  </div>
                </td>
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
                  className="items-center flex absolute w-full gap-2 px-1"
                >
                  {row.getVisibleCells().map((cell) => {
                    return (
                      <td
                        key={cell.id}
                        className="flex py-2"
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
