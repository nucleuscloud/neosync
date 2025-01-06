import { cn } from '@/libs/utils';
import { flexRender, Table } from '@tanstack/react-table';
import { useVirtualizer } from '@tanstack/react-virtual';
import { ReactElement, useRef } from 'react';
import {
  StickyHeaderTable,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from '../ui/table';
import MemoizedRow from './MemoizedRow';

interface Props<TData> {
  table: Table<TData>;

  estimateRowSize?(): number;
  rowOverscan?: number;
}

export default function FastTable<TData>(props: Props<TData>): ReactElement {
  const { table, estimateRowSize = () => 53, rowOverscan = 50 } = props;

  const { rows } = table.getRowModel();
  const tableContainerRef = useRef<HTMLDivElement>(null);

  const rowVirtualizer = useVirtualizer({
    count: rows.length,
    getScrollElement() {
      return tableContainerRef.current;
    },
    estimateSize: estimateRowSize,
    overscan: rowOverscan,
  });

  return (
    <div
      className={cn(
        'rounded-md border min-h-[145px] max-h-[1000px] relative border-gray-300 dark:border-gray-700 overflow-hidden',
        rows.length > 0 && 'overflow-auto'
      )}
      ref={tableContainerRef}
    >
      <StickyHeaderTable>
        <TableHeader className="bg-gray-100 dark:bg-gray-800 sticky top-0 z-10 px-2 grid">
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow
              key={headerGroup.id}
              className="flex flex-row items-center justify-between"
            >
              {headerGroup.headers.map((header) => {
                return (
                  <TableHead
                    key={header.id}
                    style={{
                      width:
                        header.column.columnDef.id != 'isSelected'
                          ? '187px'
                          : '20px',
                    }}
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
          className="grid relative"
          style={{ height: `${rowVirtualizer.getTotalSize()}px` }}
        >
          {rowVirtualizer.getVirtualItems().map((virtualRow) => {
            const row = rows[virtualRow.index];
            return (
              <MemoizedRow
                key={row.id}
                row={row}
                virtualRow={virtualRow}
                selected={row.getIsSelected()} // must be memoized here since row.getIsSelected() changes in place
              />
            );
          })}
        </TableBody>
      </StickyHeaderTable>
    </div>
  );
}
