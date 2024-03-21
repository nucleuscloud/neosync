import {
  StickyHeaderTable,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { cn } from '@/libs/utils';
import {
  ColumnDef,
  OnChangeFn,
  RowSelectionState,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import { useVirtualizer } from '@tanstack/react-virtual';
import { ReactElement, useRef } from 'react';
import { Mode } from '../DualListBox/columns';

interface Props<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  rowSelection: RowSelectionState;
  onRowSelectionChange: OnChangeFn<RowSelectionState>;
  tableContainerClassName?: string;
  mode?: Mode;
}

export default function ListBox<TData, TValue>(
  props: Props<TData, TValue>
): ReactElement {
  const {
    columns,
    data,
    rowSelection,
    onRowSelectionChange,
    tableContainerClassName,
    mode = 'many',
  } = props;
  const table = useReactTable({
    data,
    columns,
    state: {
      rowSelection: rowSelection,
    },
    enableRowSelection: true,
    enableMultiRowSelection: mode === 'many',
    onRowSelectionChange: onRowSelectionChange,
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
    measureElement:
      typeof window !== 'undefined' &&
      navigator.userAgent.indexOf('Firefox') === -1
        ? (element) => element?.getBoundingClientRect().height
        : undefined,
    overscan: 5,
  });

  return (
    <div
      className={cn(
        'max-h-[150px] overflow-auto relative w-full',
        tableContainerClassName
      )}
      ref={tableContainerRef}
    >
      <StickyHeaderTable>
        <TableHeader className="bg-gray-100 dark:bg-gray-800 sticky top-0 z-10 flex w-full px-2">
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow
              key={headerGroup.id}
              className="flex-none custom:flex items-center flex-row w-full"
              id="table-header-row"
            >
              {headerGroup.headers.map((header) => {
                return (
                  <TableHead
                    className="flex items-center"
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
          style={{
            height: `${rowVirtualizer.getTotalSize()}px`, // tells scrollbar how big the table is
          }}
          className="relative grid"
        >
          {rowVirtualizer.getVirtualItems().map((virtualRow) => {
            const row = rows[virtualRow.index];
            return (
              <TableRow
                data-index={virtualRow.index} // needed for dynamic row height measurement
                ref={(node) => rowVirtualizer.measureElement(node)} // measure dynamic row height
                key={row.id}
                style={{
                  transform: `translateY(${virtualRow.start}px)`,
                }}
                className="items-center flex absolute w-full px-2"
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
                      <div className="truncate">
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
