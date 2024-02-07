'use client';

import {
  ColumnDef,
  Row,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getExpandedRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import * as React from 'react';

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { formatDateTimeMilliseconds } from '@/util/util';
import { JobRunEvent } from '@neosync/sdk';
import { DataTablePagination } from './data-table-pagination';

interface DataTableProps {
  columns: ColumnDef<JobRunEvent>[];
  data: JobRunEvent[];
  isError: boolean;
}

export function DataTable({ columns, data, isError }: DataTableProps) {
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({ error: isError, id: false });
  const [sorting, setSorting] = React.useState<SortingState>([]);
  React.useEffect(() => {
    setColumnVisibility({ error: isError });
  }, [isError]);

  const [pagination, setPagination] = React.useState<number>(0);
  const [pageSize, setPageSize] = React.useState<number>(10);

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      columnVisibility,
      pagination: { pageIndex: pagination, pageSize: pageSize },
    },
    getRowCanExpand: () => true,
    onSortingChange: setSorting,
    onColumnVisibilityChange: setColumnVisibility,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getExpandedRowModel: getExpandedRowModel(),
  });

  return (
    <div className="space-y-2 rounded-md border overflow-hidden dark:border-gray-700">
      <div>
        <div className="rounded-md border overflow-hidden dark:border-gray-700 ">
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
                              header.column.columnDef.header,
                              header.getContext()
                            )}
                      </TableHead>
                    );
                  })}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {table.getRowModel().rows?.length ? (
                table.getRowModel().rows.map((row) => {
                  return (
                    <React.Fragment key={row.id}>
                      <TableRow
                        data-state={row.getIsSelected() && 'selected'}
                        onClick={row.getToggleExpandedHandler()}
                      >
                        {row.getVisibleCells().map((cell) => (
                          <TableCell key={cell.id}>
                            {flexRender(
                              cell.column.columnDef.cell,
                              cell.getContext()
                            )}
                          </TableCell>
                        ))}
                      </TableRow>
                      {row.getIsExpanded() && (
                        <TableRow>
                          <TableCell colSpan={row.getVisibleCells().length}>
                            {renderSubComponent(row)}
                          </TableCell>
                        </TableRow>
                      )}
                    </React.Fragment>
                  );
                })
              ) : (
                <TableRow>
                  <TableCell
                    colSpan={columns.length}
                    className="h-24 text-center"
                  >
                    No active runs found
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>
      <div className="pb-2">
        <DataTablePagination
          table={table}
          setPagination={setPagination}
          setPageSize={setPageSize}
        />
      </div>
    </div>
  );
}

function renderSubComponent(row: Row<JobRunEvent>): React.ReactElement {
  const isError = row.original.tasks.some((t) => t.error);
  return (
    <div className="p-5">
      <div className="rounded-md border overflow-hidden dark:border-gray-700 ">
        <Table>
          <TableHeader className="border-b dark:border-b-gray-700 bg-gray-100 dark:bg-gray-800">
            <TableRow>
              <TableHead>Id</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Time</TableHead>
              {isError && <TableHead>Error</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {row.original.tasks.map((t) => {
              return (
                <TableRow key={t.id}>
                  <TableCell>
                    <div className="flex space-x-2">
                      <span className="max-w-[500px] truncate font-medium">
                        {t.id.toString()}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex space-x-2">
                      <span className="max-w-[500px] truncate font-medium">
                        {t.type}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex space-x-2">
                      <span className="max-w-[500px] truncate font-medium">
                        {t.eventTime &&
                          formatDateTimeMilliseconds(t.eventTime.toDate())}
                      </span>
                    </div>
                  </TableCell>
                  {isError && (
                    <TableCell>
                      <div className="flex space-x-2">
                        <span className="font-medium">
                          <pre className="whitespace-pre-wrap">
                            {JSON.stringify(t.error, undefined, 2)}
                          </pre>
                        </span>
                      </div>
                    </TableCell>
                  )}
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
