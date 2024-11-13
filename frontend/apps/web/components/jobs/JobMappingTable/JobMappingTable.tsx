import {
  StickyHeaderTable,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { cn } from '@/libs/utils';
import { Transformer } from '@/shared/transformers';
import { JobMappingTransformerForm } from '@/yup-validations/jobs';
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFacetedMinMaxValues,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getSortedRowModel,
  RowData,
  useReactTable,
} from '@tanstack/react-table';
import { useVirtualizer } from '@tanstack/react-virtual';
import { ReactElement, useRef } from 'react';
import { TransformerResult } from '../SchemaTable/transformer-handler';

interface Props<TData, TValue> {
  data: TData[];
  columns: ColumnDef<TData, TValue>[];
  onTransformerUpdate(index: number, config: JobMappingTransformerForm): void;
  getAvailableTransformers(index: number): TransformerResult;
  getTransformerFromField(index: number): Transformer;
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    onTransformerUpdate(
      rowIndex: number,
      transformer: JobMappingTransformerForm
    ): void;
    getAvailableTransformers(rowIndex: number): TransformerResult;
    getTransformerFromField(index: number): Transformer;
  }
}

export default function JobMappingTable<TData, TValue>(
  props: Props<TData, TValue>
): ReactElement {
  const {
    data,
    columns,
    onTransformerUpdate,
    getAvailableTransformers,
    getTransformerFromField,
  } = props;
  const table = useReactTable({
    data,
    columns,
    initialState: {},
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getFacetedMinMaxValues: getFacetedMinMaxValues(),
    meta: {
      onTransformerUpdate: onTransformerUpdate,
      getAvailableTransformers: getAvailableTransformers,
      getTransformerFromField: getTransformerFromField,
    },
  });

  const { rows } = table.getRowModel();

  const tableContainerRef = useRef<HTMLDivElement>(null);

  const rowVirtualizer = useVirtualizer({
    count: rows.length,
    estimateSize() {
      return 33;
    },
    getScrollElement() {
      return tableContainerRef.current;
    },
    overscan: 15,
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
        <TableHeader className="sticky top-0 z-10 px-2 grid">
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow
              key={headerGroup.id}
              className="flex flex-row items-center justify-between"
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
          className="grid relative"
          style={{ height: `${rowVirtualizer.getTotalSize()}px` }}
        >
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
  );
}
