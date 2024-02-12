'use client';
import React, { ReactElement, useState } from 'react';

import './index.css';

import {
  Column,
  ColumnDef,
  ColumnFiltersState,
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
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Transformer } from '@/shared/transformers';
import { JobMappingFormValues } from '@/yup-validations/jobs';
import { SchemaTableTableToolbar } from './SchemaTableToolBar';
import { JobMapRow } from './makeData';

export type Row = JobMappingFormValues & {
  // isSelected: boolean;
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
}: DataTableProps<TData, TValue>): ReactElement {
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);

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

  // const onFilterSelect = (columnId: string, colFilters: string[]): void => {
  //   setColumnFilters((prevFilters) => {
  //     const newFilters = { ...prevFilters, [columnId]: colFilters };
  //     if (colFilters.length === 0) {
  //       delete newFilters[columnId as keyof ColumnFilters];
  //     }
  //     const filteredRows = data.filter((r) =>
  //       shouldFilterRow(r, newFilters, transformers)
  //     );
  //     setRows(filteredRows);
  //     return newFilters;
  //   });
  // };

  // const uniqueFilters = useMemo(
  //   () => getUniqueFilters(data, columnFilters, transformers),
  //   [data, columnFilters]
  // );

  return (
    <div className="app">
      All rows: ({data.length} rows) -- Currently showing: (
      {table.getRowModel().rows.length})
      <div className="pb-6">
        <SchemaTableTableToolbar table={table} />
      </div>
      <div
        className="container rounded-md border max-h-[500px] relative overflow-auto"
        ref={tableContainerRef}
      >
        <Table>
          {/* this should work to make the table sticky but isn't working here: https://uxdesign.cc/position-stuck-96c9f55d9526 */}
          <TableHeader className="bg-gray-100 dark:bg-gray-800">
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      key={header.id}
                      className="sticky top-0 z-50 h-[50px]"
                    >
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            ({}) => (
                              // <SchemaTableColumnFilterSelect
                              //   columnId={column.id}
                              //   allColumnFilters={columnFilters}
                              //   setColumnFilters={onFilterSelect}
                              //   possibleFilters={uniqueFilters[column.id]}
                              // />
                              <div>
                                <Filter column={header.column} table={table} />
                              </div>
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

// function shouldFilterRow(
//   row: JobMapRow,
//   columnFilters: ColumnFilters,
//   transformers: Transformer[],
//   columnIdToSkip?: keyof DataRow
// ): boolean {
//   for (const key of Object.keys(columnFilters)) {
//     if (columnIdToSkip && key === columnIdToSkip) {
//       continue;
//     }
//     const filters = columnFilters[key as keyof ColumnFilters];
//     if (filters.length === 0) {
//       continue;
//     }
//     switch (key) {
//       case 'transformer': {
//         const rowVal = row[key as keyof JobMapRow] as JobMappingTransformerForm;
//         if (rowVal.source === 'custom') {
//           const udfId = (rowVal.config.value as UserDefinedTransformerConfig)
//             .id;
//           const value =
//             transformers.find(
//               (t) => isUserDefinedTransformer(t) && t.id === udfId
//             )?.name ?? 'unknown transformer';
//           if (!filters.includes(value)) {
//             return false;
//           }
//         } else {
//           const value =
//             transformers.find(
//               (t) => isSystemTransformer(t) && t.source === rowVal.source
//             )?.name ?? 'unknown transformer';
//           if (!filters.includes(value)) {
//             return false;
//           }
//         }
//         break;
//       }
//       default: {
//         const value = row[key as keyof DataRow] as string;
//         if (!filters.includes(value)) {
//           return false;
//         }
//       }
//     }
//   }
//   return true;
// }

// function getUniqueFilters(
//   allRows: JobMapRow[],
//   columnFilters: ColumnFilters,
//   transformers: Transformer[]
// ): Record<string, string[]> {
//   const filterSet = {
//     schema: new Set<string>(),
//     table: new Set<string>(),
//     column: new Set<string>(),
//     dataType: new Set<string>(),
//     transformer: new Set<string>(),
//   };
//   allRows.forEach((row) => {
//     if (shouldFilterRow(row, columnFilters, transformers, 'schema')) {
//       filterSet.schema.add(row.schema);
//     }
//     if (shouldFilterRow(row, columnFilters, transformers, 'table')) {
//       filterSet.table.add(row.table);
//     }
//     if (shouldFilterRow(row, columnFilters, transformers, 'column')) {
//       filterSet.column.add(row.column);
//     }
//     if (shouldFilterRow(row, columnFilters, transformers, 'dataType')) {
//       filterSet.dataType.add(row.dataType);
//     }
//     if (shouldFilterRow(row, columnFilters, transformers, 'transformer')) {
//       filterSet.transformer.add(getTransformerFilterValue(row, transformers));
//     }
//   });
//   const uniqueColFilters: Record<string, string[]> = {};
//   Object.entries(filterSet).forEach(([key, val]) => {
//     uniqueColFilters[key] = Array.from(val).sort();
//   });
//   return uniqueColFilters;
// }

// function getTransformerFilterValue(
//   row: JobMapRow,
//   transformers: Transformer[]
// ): string {
//   if (row.transformer.value === 'custom') {
//     const udfId = (row.transformer.config as UserDefinedTransformerConfig).id;
//     return (
//       transformers.find((t) => isUserDefinedTransformer(t) && t.id === udfId)
//         ?.name ?? 'unknown transformer'
//     );
//   } else {
//     return (
//       transformers.find(
//         (t) => isSystemTransformer(t) && t.source === row.transformer.value
//       )?.name ?? 'unknown transformer'
//     );
//   }
// }

function Filter({
  column,
  table,
}: {
  column: Column<any, any>; // eslint-disable-line
  table: Table<any>; // eslint-disable-line
}) {
  const firstValue = table
    .getPreFilteredRowModel()
    .flatRows[0]?.getValue(column.id);

  return typeof firstValue === 'number' ? (
    <div className="flex space-x-2">
      <input
        type="number"
        value={((column.getFilterValue() as any)?.[0] ?? '') as string} // eslint-disable-line
        onChange={
          (e) => column.setFilterValue((old: any) => [e.target.value, old?.[1]]) // eslint-disable-line
        }
        placeholder={`Min`}
        className="w-24 border shadow rounded"
      />
      <input
        type="number"
        value={((column.getFilterValue() as any)?.[1] ?? '') as string} // eslint-disable-line
        onChange={
          (e) => column.setFilterValue((old: any) => [old?.[0], e.target.value]) // eslint-disable-line
        }
        placeholder={`Max`}
        className="w-24 border shadow rounded"
      />
    </div>
  ) : (
    <input
      type="text"
      value={(column.getFilterValue() ?? '') as string}
      onChange={(e) => column.setFilterValue(e.target.value)}
      placeholder={`Search...`}
      className="w-36 border shadow rounded"
    />
  );
}
