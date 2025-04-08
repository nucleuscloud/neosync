'use client';

import {
  ColumnDef,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import * as React from 'react';

import { DataTablePagination } from '@/components/table/data-table-pagination';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  JobRunEvent,
  JobRunEventMetadata,
  JobRunSyncMetadata,
} from '@neosync/sdk';
import { useLocalStorage } from 'usehooks-ts';

interface DataTableProps {
  columns: ColumnDef<JobRunEvent>[];
  data: JobRunEvent[];
  isError: boolean;
  onViewSelectClicked(schema: string, table: string): void;
}

export function DataTable({ columns, data, isError }: DataTableProps) {
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({ error: isError, id: false });
  const [sorting, setSorting] = React.useState<SortingState>([]);
  React.useEffect(() => {
    setColumnVisibility({ error: isError });
  }, [isError]);

  const [pagination, setPagination] = React.useState<number>(0);
  const [pageSize, setPageSize] = useLocalStorage<number>(
    'job-run-activity-table-page-size',
    10
  );

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
  });

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-row items-center justify-between">
        <div className="text-xl font-semibold">Activity Table</div>
      </div>

      <div className="space-y-2 rounded-md border overflow-hidden dark:border-gray-700">
        <Table>
          <TableHeader className="bg-gray-100 dark:bg-gray-800">
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id} className="pl-2">
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
                    <TableRow data-state={row.getIsSelected() && 'selected'}>
                      {row.getVisibleCells().map((cell) => (
                        <TableCell key={cell.id}>
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext()
                          )}
                        </TableCell>
                      ))}
                    </TableRow>
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
      <DataTablePagination
        table={table}
        setPagination={setPagination}
        setPageSize={setPageSize}
      />
    </div>
  );
}

export function getJobSyncMetadata(
  metadata?: JobRunEventMetadata
): JobRunSyncMetadata | null {
  if (metadata?.metadata.case === 'syncMetadata') {
    const md = metadata.metadata.value;
    if (md.schema && md.table) {
      return md;
    }
  }
  return null;
}
