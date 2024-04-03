'use client';
import React, { ReactElement } from 'react';

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

import { CardDescription, CardTitle } from '@/components/ui/card';
import {
  StickyHeaderTable,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { JobMappingFormValues } from '@/yup-validations/jobs';
import { GoWorkflow } from 'react-icons/go';
import { SchemaTableToolbar } from './SchemaTableToolBar';
import { JobType, SchemaConstraintHandler } from './schema-constraint-handler';
import { TransformerHandler } from './transformer-handler';

export type Row = JobMappingFormValues;

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  transformerHandler: TransformerHandler;
  constraintHandler: SchemaConstraintHandler;
  jobType: JobType;
}

export default function SchemaPageTable<TData, TValue>({
  columns,
  data,
  transformerHandler,
  constraintHandler,
  jobType,
}: DataTableProps<TData, TValue>): ReactElement {
  const table = useReactTable({
    data,
    columns,
    initialState: {
      sorting: [
        { id: 'schema', desc: true },
        { id: 'table', desc: true },
      ],
      columnVisibility: {
        schema: false,
        table: false,
      },
    },
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getFacetedMinMaxValues: getFacetedMinMaxValues(),
  });

  const { rows } = table.getRowModel();

  const tableContainerRef = React.useRef<HTMLDivElement>(null);

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
      <div className="flex flex-row items-center gap-2 pt-4">
        <div className="flex">
          <GoWorkflow className="h-4 w-4" />
        </div>
        <CardTitle>Transformer Mapping</CardTitle>
      </div>
      <CardDescription className="pt-2">
        Map Transformers to every column below.
      </CardDescription>
      <div className="z-50 pt-4">
        <SchemaTableToolbar
          table={table}
          transformerHandler={transformerHandler}
          constraintHandler={constraintHandler}
          jobType={jobType}
        />
      </div>
      <div
        className="rounded-md border max-h-[500px] relative overflow-x-auto"
        ref={tableContainerRef}
      >
        <StickyHeaderTable>
          <TableHeader className="bg-gray-100 dark:bg-gray-800 sticky top-0 z-10 flex w-full px-2">
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow
                key={headerGroup.id}
                className="flex flex-row items-center justify-between w-full"
                id="table-header-row"
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
            style={{
              height: `${rowVirtualizer.getTotalSize()}px`, //tells scrollbar how big the table is
            }}
          >
            {rows.length === 0 && (
              <TableRow className="flex justify-center items-center py-10 text-gray-500">
                <td>No Schema(s) or Table(s) selected.</td>
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
                        <div>
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext()
                          )}
                        </div>
                      </td>
                    );
                  })}
                </TableRow>
              );
            })}
          </TableBody>
        </StickyHeaderTable>
      </div>
      <div className="text-xs text-gray-600 dark:text-300 pt-4">
        Total rows: ({new Intl.NumberFormat('en-US').format(data.length)}) Rows
        visible: (
        {new Intl.NumberFormat('en-US').format(table.getRowModel().rows.length)}
        )
      </div>
    </div>
  );
}
