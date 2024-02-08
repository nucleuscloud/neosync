'use client';
import React, { ReactElement, useState } from 'react';

import './index.css';

import {
  ColumnDef,
  ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFacetedMinMaxValues,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getSortedRowModel,
  Row,
  useReactTable,
} from '@tanstack/react-table';

import { useVirtualizer } from '@tanstack/react-virtual';

import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { JobMapRow, makeData } from './makeData';
import { SchemaTableColumnHeader } from './SchemaTableColumnHeader';
import { SchemaTableTableToolbar } from './SchemaTableToolBar';

interface Props {
  columns: ColumnDef<JobMapRow>[];
}

export default function SchemaTableTest(props: Props): ReactElement {
  const { columns } = props;
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );
  const [data, _setData] = useState(() => makeData(10000, 10, 100));

  const table = useReactTable({
    data,
    columns,
    state: {
      columnFilters,
    },
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getFacetedMinMaxValues: getFacetedMinMaxValues(),
    debugTable: true,
  });

  const { rows } = table.getRowModel();

  //The virtualizer needs to know the scrollable container elementS
  const tableContainerRef = React.useRef<HTMLDivElement>(null);

  const rowVirtualizer = useVirtualizer({
    count: rows.length,
    estimateSize: () => 33, //estimate row height for accurate scrollbar dragging
    getScrollElement: () => tableContainerRef.current,
    //measure dynamic row height, except in firefox because it measures table border height incorrectly
    measureElement:
      typeof window !== 'undefined' &&
      navigator.userAgent.indexOf('Firefox') === -1
        ? (element) => element?.getBoundingClientRect().height
        : undefined,
    overscan: 5,
  });

  //All important CSS styles are included as inline styles for this example. This is not recommended for your code.
  return (
    <div className="app">
      All rows: ({data.length} rows) -- Currently showing: (
      {table.getRowModel().rows.length})
      <div className="pb-6">
        <SchemaTableTableToolbar table={table} />
      </div>
      <div
        className="container rounded-md border"
        ref={tableContainerRef}
        style={{
          overflow: 'auto', //our scrollable table container
          position: 'relative', //needed for sticky header
          height: '500px', //should be a fixed height
          width: '100%',
        }}
      >
        {/* Even though we're still using sematic table tags, we must use CSS grid and flexbox for dynamic row heights */}
        <Table>
          <TableHeader className="bg-gray-100 dark:bg-gray-800">
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            ({ column }) => (
                              <SchemaTableColumnHeader
                                column={column}
                                title={column.id}
                                table={table}
                                header={header}
                                // data={table
                                //   .getRowModel()
                                //   .rows.map((row) => row.getValue(column.id))}
                              />
                            ),
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
              display: 'grid',
              height: `${rowVirtualizer.getTotalSize()}px`, //tells scrollbar how big the table is
              position: 'relative', //needed for absolute positioning of rows
            }}
          >
            {rowVirtualizer.getVirtualItems().map((virtualRow) => {
              const row = rows[virtualRow.index] as Row<JobMapRow>;
              return (
                <TableRow
                  data-index={virtualRow.index} //needed for dynamic row height measurement
                  ref={(node) => rowVirtualizer.measureElement(node)} //measure dynamic row height
                  key={row.id}
                  style={{
                    display: 'flex',
                    position: 'absolute', // must be absolute or hits unlimited renders
                    transform: `translateY(${virtualRow.start}px)`, //this should always be a `style` as it changes on scroll
                  }}
                >
                  {row.getVisibleCells().map((cell) => {
                    return (
                      <td
                        key={cell.id}
                        style={{
                          display: 'flex',
                          width: cell.column.getSize(),
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
        </Table>
      </div>
    </div>
  );
}
