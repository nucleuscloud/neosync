'use client';

import {
  ColumnDef,
  ColumnFiltersState,
  Row,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import * as React from 'react';

import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Avatar, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useToast } from '@/components/ui/use-toast';
import { useGetAccountMembers } from '@/libs/hooks/useGetAccountMembers';
import { getErrorMessage } from '@/util/util';
import { PlainMessage } from '@bufbuild/protobuf';
import { AccountUser } from '@neosync/sdk';
import { DotsHorizontalIcon } from '@radix-ui/react-icons';

interface ColumnProps {
  onDeleted(id: string): void;
  accountId: string;
}

function getColumns(
  props: ColumnProps
): ColumnDef<PlainMessage<AccountUser>>[] {
  const { onDeleted, accountId } = props;
  return [
    {
      accessorKey: 'name',
      header: 'Name',
      cell: ({ row }) => (
        <div className="flex flex-row items-center gap-4">
          <Avatar className="mr-2 h-12 w-12">
            <AvatarImage
              src={
                row.original.image ||
                `https://avatar.vercel.sh/${row.getValue('name')}.png`
              }
              alt={row.getValue('name')}
            />
          </Avatar>
          <span className=" truncate font-medium">{row.getValue('name')}</span>
        </div>
      ),
    },
    {
      accessorKey: 'email',
      header: 'Email',
      cell: ({ row }) => (
        <span className=" truncate font-medium">{row.getValue('email')}</span>
      ),
    },
    {
      id: 'actions',
      cell: ({ row }) => (
        <DataTableRowActions
          accountId={accountId}
          row={row}
          onDeleted={() => onDeleted(row.id)}
        />
      ),
    },
  ];
}

interface Props {
  accountId: string;
}

export default function MembersTable(props: Props) {
  const { accountId } = props;
  const { data, isLoading, mutate } = useGetAccountMembers(accountId || '');
  if (isLoading) {
    return <SkeletonTable />;
  }
  return (
    <DataTable
      data={data?.users || []}
      accountId={accountId}
      onDeleted={() => mutate()}
    />
  );
}

interface DataTableProps {
  data: AccountUser[];
  accountId: string;
  onDeleted(id: string): void;
}

function DataTable(props: DataTableProps): React.ReactElement {
  const { data, accountId, onDeleted } = props;
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = React.useState({});

  const columns = getColumns({ accountId, onDeleted });

  const table = useReactTable({
    data,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
    },
  });

  return (
    <div className="w-full">
      <div className="flex items-center py-4">
        <Input
          placeholder="Filter emails..."
          value={(table.getColumn('email')?.getFilterValue() as string) ?? ''}
          onChange={(event) =>
            table.getColumn('email')?.setFilterValue(event.target.value)
          }
          className="max-w-sm"
        />
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
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
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && 'selected'}
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
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-end space-x-2 py-4">
        <div className="flex-1 text-sm text-muted-foreground">
          {table.getFilteredSelectedRowModel().rows.length} of{' '}
          {table.getFilteredRowModel().rows.length} row(s) selected.
        </div>
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}

interface DataTableRowActionsProps<TData> {
  row: Row<TData>;
  onDeleted(): void;
  accountId: string;
}

function DataTableRowActions<TData>({
  row,
  onDeleted,
  accountId,
}: DataTableRowActionsProps<TData>) {
  const user = row.original as AccountUser;
  const { toast } = useToast();

  async function onRemove(): Promise<void> {
    try {
      await removeUserFromTeamAccount(accountId, user.id);
      toast({
        title: 'User removed successfully!',
      });
      onDeleted();
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to remove user from account',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <DropdownMenu
      modal={false} // needed because otherwise this breaks after a single use in conjunction with the delete dialog
    >
      <DropdownMenuTrigger className="hover:bg-gray-100 dark:hover:bg-gray-800 py-1 px-2 rounded-lg">
        <DotsHorizontalIcon className="h-4 w-4" />
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DeleteConfirmationDialog
          trigger={
            <DropdownMenuItem
              className="cursor-pointer"
              onSelect={(e) => e.preventDefault()} // needed for the delete modal to not automatically close
            >
              Remove
            </DropdownMenuItem>
          }
          headerText="Are you sure you want to remove this user?"
          description=""
          onConfirm={() => onRemove()}
        />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

async function removeUserFromTeamAccount(
  accountId: string,
  userId: string
): Promise<void> {
  const res = await fetch(
    `/api/users/accounts/${accountId}/members?id=${userId}`,
    {
      method: 'DELETE',
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}
