import {
  StickyHeaderTable,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { cn } from '@/libs/utils';
import { Table, flexRender } from '@tanstack/react-table';
import { useVirtualizer } from '@tanstack/react-virtual';
import { ReactElement, useRef } from 'react';
import { Skeleton } from '../ui/skeleton';

interface Props<TData> {
  table: Table<TData>;
  tableContainerClassName?: string;
  isDataLoading?: boolean;
  noDataMessage?: string;
}

export default function ListBox<TData>(props: Props<TData>): ReactElement {
  const { table, tableContainerClassName, isDataLoading, noDataMessage } =
    props;
  const { rows } = table.getRowModel();
  const tableContainerRef = useRef<HTMLDivElement>(null);

  const rowVirtualizer = useVirtualizer({
    count: rows.length,
    estimateSize: () => 33,
    getScrollElement: () => tableContainerRef.current,
    overscan: 5,
  });

  if (isDataLoading) {
    return <Skeleton className="w-full h-full" />;
  }

  return (
    <div
      className={cn(
        'h-[164px]  rounded-md border border-gray-300 dark:border-gray-700 w-[350px]',
        tableContainerClassName,
        rows.length > 0 && 'overflow-auto'
      )}
      ref={tableContainerRef}
    >
      <StickyHeaderTable className="w-full">
        <TableHeader className="bg-gray-100 dark:bg-gray-800 sticky top-0 z-10 grid ">
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow
              key={headerGroup.id}
              className="flex flex-row px-2 w-full"
            >
              {headerGroup.headers.map((header) => {
                return (
                  <TableHead
                    className="flex items-center"
                    id="table-head"
                    key={header.id}
                    style={{ minWidth: `${header.column.getSize()}px` }}
                    colSpan={header.colSpan}
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
            height: `${rowVirtualizer.getTotalSize()}px`, // tells scrollbar how big the table is
          }}
        >
          {rows.length === 0 && !!noDataMessage && (
            <TableRow className="flex justify-center items-center py-10 text-gray-500">
              <td className="px-4">{noDataMessage}</td>
            </TableRow>
          )}
          {rowVirtualizer.getVirtualItems().map((virtualRow) => {
            const row = rows[virtualRow.index];
            return (
              <TableRow
                data-index={virtualRow.index} // needed for dynamic row height measurement
                ref={(node) => {
                  rowVirtualizer.measureElement(node);
                }} // measure dynamic row height
                key={row.id}
                className="items-center flex  w-full px-2"
                onClick={row.getToggleSelectedHandler()}
              >
                {row.getVisibleCells().map((cell) => {
                  return (
                    <TableCell
                      className="px-0"
                      key={cell.id}
                      style={{
                        minWidth: cell.column.getSize(),
                      }}
                    >
                      <div>
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </div>
                    </TableCell>
                  );
                })}
              </TableRow>
            );
          })}
        </TableBody>
      </StickyHeaderTable>
    </div>
  );
}
