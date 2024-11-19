import { CardDescription, CardTitle } from '@/components/ui/card';
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
import { JobMapping } from '@neosync/sdk';
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  Row,
  RowData,
  useReactTable,
} from '@tanstack/react-table';
import { useVirtualizer } from '@tanstack/react-virtual';
import { ReactElement, useRef } from 'react';
import { GoWorkflow } from 'react-icons/go';
import { ImportMappingsConfig } from '../SchemaTable/ImportJobMappingsButton';
import { SchemaTableToolbar } from '../SchemaTable/SchemaTableToolBar';
import { TransformerResult } from '../SchemaTable/transformer-handler';

interface Props<TData, TValue> {
  data: TData[];
  columns: ColumnDef<TData, TValue>[];
  onTransformerUpdate(index: number, config: JobMappingTransformerForm): void;
  getAvailableTransformers(index: number): TransformerResult;
  getTransformerFromField(index: number): Transformer;

  onTransformerBulkUpdate(
    indices: number[],
    config: JobMappingTransformerForm
  ): void;
  getAvalableTransformersForBulk(rows: Row<TData>[]): TransformerResult;
  getTransformerFromFieldValue(value: JobMappingTransformerForm): Transformer;

  isApplyDefaultTransformerButtonDisabled: boolean;
  onApplyDefaultClick(override: boolean): void;

  onExportMappingsClick(selected: Row<TData>[], shouldFormat: boolean): void;
  onImportMappingsClick(
    jobmappings: JobMapping[],
    config: ImportMappingsConfig
  ): void;
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
    onExportMappingsClick,
    onImportMappingsClick,
    getAvalableTransformersForBulk,
    getTransformerFromFieldValue,
    isApplyDefaultTransformerButtonDisabled,
    onApplyDefaultClick,
    onTransformerBulkUpdate,
  } = props;

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    meta: {
      onTransformerUpdate: onTransformerUpdate,
      getAvailableTransformers: getAvailableTransformers,
      getTransformerFromField: getTransformerFromField,
    },
    debugAll: true,
  });

  const { rows } = table.getRowModel();

  const tableContainerRef = useRef<HTMLDivElement>(null);

  const rowVirtualizer = useVirtualizer({
    count: rows.length,
    estimateSize() {
      return 53;
    },
    getScrollElement() {
      return tableContainerRef.current;
    },
    overscan: 15,
  });

  return (
    <div>
      <div className="flex flex-row items-center gap-2 pt-4 ">
        <div className="flex">
          <GoWorkflow className="h-4 w-4" />
        </div>
        <CardTitle>Transformer Mapping</CardTitle>
      </div>
      <CardDescription className="pt-2">
        Map Transformers to every column below.
      </CardDescription>
      <div className="z-50 pt-4">
        <SchemaTableToolbar<TData>
          table={table}
          displayApplyDefaultTransformersButton={true}
          isApplyDefaultButtonDisabled={isApplyDefaultTransformerButtonDisabled}
          getAllowedTransformers={getAvalableTransformersForBulk}
          getTransformerFromField={getTransformerFromFieldValue}
          onApplyDefaultClick={onApplyDefaultClick}
          onBulkUpdate={onTransformerBulkUpdate}
          onExportMappingsClick={(shouldFormat) =>
            onExportMappingsClick(
              table.getSelectedRowModel().rows,
              shouldFormat
            )
          }
          onImportMappingsClick={onImportMappingsClick}
        />
      </div>

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
                  key={row.id}
                  style={{
                    transform: `translateY(${virtualRow.start}px)`,
                    height: `${virtualRow.size}px`,
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
