import {
  StickyHeaderTable,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
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

interface Props<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  rowSelection: RowSelectionState;
  onRowSelectionChange: OnChangeFn<RowSelectionState>;
}

export default function ListBox<TData, TValue>(
  props: Props<TData, TValue>
): ReactElement {
  const { columns, data, rowSelection, onRowSelectionChange } = props;
  const table = useReactTable({
    data,
    columns,
    state: {
      rowSelection: rowSelection,
    },
    enableRowSelection: true,
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
    <div className="w-full" ref={tableContainerRef}>
      <StickyHeaderTable>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow
              key={headerGroup.id}
              className="flex items-center flex-row w-full"
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
                className="items-center flex absolute w-full"
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
