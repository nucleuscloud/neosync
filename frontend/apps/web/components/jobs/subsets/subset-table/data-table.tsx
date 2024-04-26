'use client';

import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';

import ButtonText from '@/components/ButtonText';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Cross2Icon } from '@radix-ui/react-icons';
interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
}

export function DataTable<TData, TValue>({
  columns,
  data,
}: DataTableProps<TData, TValue>) {
  const table = useReactTable({
    data,
    columns,
    enableRowSelection: true,
    initialState: {
      columnVisibility: {
        schema: false,
        table: false,
      },
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
  });

  const { rows } = table.getRowModel();

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-row items-center gap-2 justify-end">
        <div className="flex flex-row items-center gap-2">
          <Button
            disabled={table.getState().columnFilters.length === 0}
            type="button"
            variant="outline"
            className="px-2 lg:px-3"
            onClick={() => table.resetColumnFilters()}
          >
            <ButtonText
              leftIcon={<Cross2Icon className="h-3 w-3" />}
              text="Clear filters"
            />
          </Button>
        </div>
      </div>
      <div className="rounded-md border overflow-hidden dark:border-gray-700">
        <Table className="table-fixed">
          <TableHeader className="bg-gray-100 dark:bg-gray-800">
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead
                      key={header.id}
                      className={
                        header.id === 'select'
                          ? 'w-[44px] pl-2 '
                          : 'w-[197px] px-0'
                      }
                    >
                      <div className="flex flex-row">
                        {header.isPlaceholder
                          ? null
                          : flexRender(
                              header.column.columnDef.header,
                              header.getContext()
                            )}
                        {header.column.getCanFilter() ? <div></div> : null}
                      </div>
                    </TableHead>
                  );
                })}
              </TableRow>
            ))}
          </TableHeader>
        </Table>
        <ScrollArea className="max-h-[700px] overflow-y-auto">
          <Table className="table-fixed">
            <TableBody>
              {rows.length === 0 && (
                <TableRow className="flex justify-center items-center py-10 text-gray-500">
                  <td>No schema(s) or table(s) found.</td>
                </TableRow>
              )}
              {rows.map((row) => {
                return (
                  <TableRow
                    key={row.id}
                    data-state={row.getIsSelected() && 'selected'}
                  >
                    {row.getVisibleCells().map((cell) => {
                      return (
                        <TableCell
                          className={
                            cell.column.id === 'select'
                              ? ' w-[40px]'
                              : 'w-[197px]'
                          }
                          key={cell.id}
                        >
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext()
                          )}
                        </TableCell>
                      );
                    })}
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </ScrollArea>
      </div>
    </div>
  );
}
