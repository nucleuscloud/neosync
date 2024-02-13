'use client';
import React, { ReactElement, useState } from 'react';

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

import {
  StickyHeaderTable,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Transformer } from '@/shared/transformers';
import {
  JobMappingFormValues,
  JobMappingTransformerForm,
  SchemaFormValues,
} from '@/yup-validations/jobs';
import { SchemaTableToolbar } from './SchemaTableToolBar';

import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { useFormContext } from 'react-hook-form';
import TransformerSelect from '../SchemaTable/TransformerSelect';

export type Row = JobMappingFormValues & {
  formIdx: number;
};

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  transformers: Transformer[];
}

export default function SchemaTableTest<TData, TValue>({
  columns,
  data,
  transformers,
}: DataTableProps<TData, TValue>): ReactElement {
  const [bulkTransformer, setBulkTransformer] =
    useState<JobMappingTransformerForm>({
      source: '',
      config: { case: '', value: {} },
    });
  // const [columnVisibility, setColumnVisibility] =
  //   React.useState<VisibilityState>({ schema: false });

  const table = useReactTable({
    data,
    columns,
    state: {
      // columnVisibility,
    },

    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getFacetedMinMaxValues: getFacetedMinMaxValues(),
    // onColumnVisibilityChange: setColumnVisibility,
    debugTable: true,
  });

  const { rows } = table.getRowModel();

  const tableContainerRef = React.useRef<HTMLDivElement>(null);

  const form = useFormContext<SingleTableSchemaFormValues | SchemaFormValues>();

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
    <div>
      <div className="pb-6 z-50 flex flex-col gap-2">
        <div className="text-sm pb-10 ">
          Use the Schema Table below to map your columns to Transformers.{' '}
        </div>
        <SchemaTableToolbar table={table} data={data} />
        <div className="w-[250px]">
          <TransformerSelect
            transformers={transformers}
            value={bulkTransformer}
            side={'bottom'}
            onSelect={(value) => {
              table.getSelectedRowModel().rows.forEach((r) => {
                form.setValue(`mappings.${r.index}.transformer`, value, {
                  shouldDirty: true,
                });
              });
              setBulkTransformer({
                source: '',
                config: { case: '', value: {} },
              });
            }}
            placeholder="Bulk update Transformers..."
          />
        </div>
      </div>
      <div
        className="rounded-md border max-h-[500px] relative overflow-auto"
        ref={tableContainerRef}
      >
        <StickyHeaderTable>
          <TableHeader className="bg-gray-100 dark:bg-gray-800 sticky top-0 z-10 flex justify-start px-6">
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      key={header.id}
                      style={{ width: header.getSize() }}
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
            className="relative, grid"
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
                  className="items-center flex absolute"
                >
                  {row.getVisibleCells().map((cell) => {
                    return (
                      <td
                        key={cell.id}
                        className="flex items-start px-6 py-2"
                        style={{ width: cell.column.columnDef.size }}
                      >
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <div className="truncate">
                                {flexRender(
                                  cell.column.columnDef.cell,
                                  cell.getContext()
                                )}
                              </div>
                            </TooltipTrigger>
                            <TooltipContent
                              className="w-auto p-2 z-50"
                              side="right"
                            >
                              {getCellDisplayValue(cell.getValue())}
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
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

const getCellDisplayValue = (value: unknown): string => {
  if (
    typeof value === 'object' &&
    value !== null &&
    'source' in value &&
    value != undefined
  ) {
    return (value as { source: unknown }).source as string;
  }
  return String(value);
};
