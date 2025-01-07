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
  /**
   * The table instance to render.
   */
  table: Table<TData>;
  /**
   * This function is used to estimate the height of each row.
   * It should be the height of your row including padding and any other styling.
   * This is used to calculate the total height of the table and to smooth out scrolling.
   * If this is not set, it will default to 53px which is the height of the default row.
   */
  estimateRowSize?(): number;
  /**
   * This is the number of rows to render above and below the visible rows to help with smooth scrolling.
   * This should be set to a value that is a multiple of your estimateRowSize.
   * This is a trade off between performance and CPU.
   */
  rowOverscan?: number;
}

/**
 * This table uses memoized rows and cells to improve performance ontop of a scroll virtualizer for infinite scrolling of large datasets.
 * Depending on your column makeup, the memoized cells may need updated to handle your specific display column.
 * It's also very impportant to memoize the table dataset so that it is stable and only changes when needed.
 * This will maximize performance of the table.
 *
 * It's also very important to set your estimateRowSize function to be the correct height of your row. Otherwise it is a guess and will have to be re-calculated, which hurts performance.
 * Configuring the overscan helps smooth out scrolling, but will increase CPU. For fast systems this is very noticeable and helps with reducing white flashing during quick scrolls.
 */
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
