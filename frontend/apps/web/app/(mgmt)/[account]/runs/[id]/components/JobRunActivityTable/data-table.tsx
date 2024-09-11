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

import ButtonText from '@/components/ButtonText';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { formatDateTimeMilliseconds } from '@/util/util';
import {
  JobRunEvent,
  JobRunEventMetadata,
  JobRunSyncMetadata,
} from '@neosync/sdk';
import { DataTablePagination } from './data-table-pagination';

interface DataTableProps {
  columns: ColumnDef<JobRunEvent>[];
  data: JobRunEvent[];
  isError: boolean;
  onViewSelectClicked(schema: string, table: string): void;
}

export function DataTable({
  columns,
  data,
  isError,
  onViewSelectClicked,
}: DataTableProps) {
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
                            <RunEventSubTable
                              row={row}
                              onViewSelectClicked={onViewSelectClicked}
                            />
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

interface RunEventSubTableProps {
  row: Row<JobRunEvent>;
  onViewSelectClicked(schema: string, table: string): void;
}

function RunEventSubTable(props: RunEventSubTableProps): React.ReactElement {
  const { row, onViewSelectClicked } = props;
  const isError = row.original.tasks.some((t) => t.error);
  const syncMd = getJobSyncMetadata(row.original.metadata);

  function onSelectClicked(): void {
    if (syncMd) {
      onViewSelectClicked(syncMd.schema, syncMd.table);
    }
  }

  return (
    <div className="p-5 flex flex-col gap-2">
      {!!syncMd && (
        <div className="flex flex-col gap-2">
          <div>
            <Button type="button" onClick={() => onSelectClicked()}>
              <ButtonText text="View SELECT Query" />
            </Button>
          </div>
        </div>
      )}
      <div className="flex flex-col gap-2">
        <h2 className="tracking-tight">Event History</h2>
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
    </div>
  );
}

function getJobSyncMetadata(
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
